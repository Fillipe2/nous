package middleware

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	limiteGlobalPorMinuto   = 300
	limiteSensivelPorMinuto = 20
	intervaloRetryRedis     = time.Minute
)

type rateBucket struct {
	contador int
	resetEm  time.Time
}

type rateLimiterStore interface {
	Excedeu(ctx context.Context, chave string, limite int, janela time.Duration) (bool, error)
	Close() error
}

type memoryRateLimiter struct{}

type redisRateLimiter struct {
	client *redis.Client
}

var rateState = struct {
	sync.Mutex
	buckets map[string]rateBucket
}{buckets: make(map[string]rateBucket)}

type rateLimiterCell struct {
	store            rateLimiterStore
	redisURL         string
	proximaTentativa time.Time
}

var (
	rateLimiterCellPtr     atomic.Pointer[rateLimiterCell]
	rateLimiterReconfigMu  sync.Mutex
	rateLimiterConfigMu    sync.RWMutex
	rateLimiterRedisURL    string
	rateLimiterConfigurado bool
)

var redisRateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

func LimitarTaxa(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ClienteIP(r)
		classe := classeRota(r)
		if excedeuLimite(r.Context(), "global:"+ip, limiteGlobalPorMinuto) || (classe != "" && excedeuLimite(r.Context(), "sensivel:"+ip+":"+classe, limiteSensivelPorMinuto)) {
			w.Header().Set("Retry-After", "60")
			responderErro(w, r, http.StatusTooManyRequests, "muitas requisicoes, tente novamente em instantes")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func excedeuLimite(ctx context.Context, chave string, limite int) bool {
	excedeu, err := obterRateLimiter().Excedeu(ctx, chave, limite, time.Minute)
	if err == nil {
		return excedeu
	}
	marcarRateLimiterIndisponivel(err)
	excedeu, _ = memoryRateLimiter{}.Excedeu(ctx, chave, limite, time.Minute)
	return excedeu
}

func obterRateLimiter() rateLimiterStore {
	redisURL := redisURLConfigurada()
	if cell := rateLimiterCellPtr.Load(); cell != nil && celulaUtilizavel(cell, redisURL, time.Now()) {
		return cell.store
	}
	return reconfigurarRateLimiter(redisURL)
}

func ConfigurarRateLimiter(redisURL string) {
	redisURL = strings.TrimSpace(redisURL)
	rateLimiterConfigMu.Lock()
	rateLimiterRedisURL = redisURL
	rateLimiterConfigurado = true
	rateLimiterConfigMu.Unlock()

	rateLimiterReconfigMu.Lock()
	if cell := rateLimiterCellPtr.Load(); cell != nil && cell.redisURL != redisURL && cell.store != nil {
		_ = cell.store.Close()
		rateLimiterCellPtr.Store(nil)
	}
	rateLimiterReconfigMu.Unlock()
}

func redisURLConfigurada() string {
	rateLimiterConfigMu.RLock()
	if rateLimiterConfigurado {
		redisURL := rateLimiterRedisURL
		rateLimiterConfigMu.RUnlock()
		return redisURL
	}
	rateLimiterConfigMu.RUnlock()

	rateLimiterConfigMu.Lock()
	defer rateLimiterConfigMu.Unlock()
	if !rateLimiterConfigurado {
		rateLimiterRedisURL = strings.TrimSpace(os.Getenv("REDIS_URL"))
		rateLimiterConfigurado = true
	}
	return rateLimiterRedisURL
}

func celulaUtilizavel(cell *rateLimiterCell, redisURL string, agora time.Time) bool {
	if cell.redisURL != redisURL {
		return false
	}
	return redisURL == "" || isRedisRateLimiter(cell.store) || agora.Before(cell.proximaTentativa)
}

func reconfigurarRateLimiter(redisURL string) rateLimiterStore {
	rateLimiterReconfigMu.Lock()
	defer rateLimiterReconfigMu.Unlock()

	agora := time.Now()
	cell := rateLimiterCellPtr.Load()
	if cell != nil && celulaUtilizavel(cell, redisURL, agora) {
		return cell.store
	}
	if cell != nil && cell.redisURL != redisURL && cell.store != nil {
		_ = cell.store.Close()
	}
	novoStore := configurarRateLimiter(redisURL)
	novaCell := &rateLimiterCell{store: novoStore, redisURL: redisURL}
	if redisURL != "" && !isRedisRateLimiter(novoStore) {
		novaCell.proximaTentativa = agora.Add(intervaloRetryRedis)
	}
	rateLimiterCellPtr.Store(novaCell)
	return novoStore
}

func configurarRateLimiter(redisURL string) rateLimiterStore {
	if redisURL == "" {
		log.Printf("rate limit backend: memoria")
		return memoryRateLimiter{}
	}
	opcoes, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("REDIS_URL invalida para rate limit, usando memoria: %v", err)
		return memoryRateLimiter{}
	}
	client := redis.NewClient(opcoes)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		log.Printf("redis indisponivel para rate limit, usando memoria: %v", err)
		return memoryRateLimiter{}
	}
	log.Printf("rate limit backend: redis")
	return &redisRateLimiter{client: client}
}

func marcarRateLimiterIndisponivel(err error) {
	rateLimiterReconfigMu.Lock()
	defer rateLimiterReconfigMu.Unlock()

	cell := rateLimiterCellPtr.Load()
	if cell != nil && isRedisRateLimiter(cell.store) {
		log.Printf("rate limit backend indisponivel, usando memoria: %v", err)
		_ = cell.store.Close()
	} else if cell == nil {
		log.Printf("rate limit backend indisponivel, usando memoria: %v", err)
	}
	redisURL := ""
	if cell != nil {
		redisURL = cell.redisURL
	}
	rateLimiterCellPtr.Store(&rateLimiterCell{
		store:            memoryRateLimiter{},
		redisURL:         redisURL,
		proximaTentativa: time.Now().Add(intervaloRetryRedis),
	})
}

func isRedisRateLimiter(store rateLimiterStore) bool {
	_, ok := store.(*redisRateLimiter)
	return ok
}

func CloseRateLimiter() {
	rateLimiterReconfigMu.Lock()
	defer rateLimiterReconfigMu.Unlock()
	if cell := rateLimiterCellPtr.Load(); cell != nil && cell.store != nil {
		_ = cell.store.Close()
	}
	rateLimiterCellPtr.Store(nil)
}

func (memoryRateLimiter) Excedeu(_ context.Context, chave string, limite int, janela time.Duration) (bool, error) {
	agora := time.Now()
	rateState.Lock()
	defer rateState.Unlock()

	bucket := rateState.buckets[chave]
	if bucket.resetEm.IsZero() || agora.After(bucket.resetEm) {
		rateState.buckets[chave] = rateBucket{contador: 1, resetEm: agora.Add(janela)}
		limparBucketsExpirados(agora)
		return false, nil
	}

	bucket.contador++
	rateState.buckets[chave] = bucket
	return bucket.contador > limite, nil
}

func (memoryRateLimiter) Close() error {
	return nil
}

func (r *redisRateLimiter) Excedeu(ctx context.Context, chave string, limite int, janela time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	contador, err := redisRateLimitScript.Run(ctx, r.client, []string{chave}, janela.Milliseconds()).Int64()
	if err != nil {
		return false, err
	}
	return contador > int64(limite), nil
}

func (r *redisRateLimiter) Close() error {
	return r.client.Close()
}

func limparBucketsExpirados(agora time.Time) {
	for chave, bucket := range rateState.buckets {
		if agora.After(bucket.resetEm) {
			delete(rateState.buckets, chave)
		}
	}
}

func classeRota(r *http.Request) string {
	path := r.URL.Path
	if r.Method == http.MethodPost && (path == "/auth/login" || path == "/auth/refresh" || path == "/auth/esqueci-senha" || path == "/auth/redefinir-senha" || path == "/usuarios" || path == "/leads") {
		return path
	}
	if strings.HasPrefix(path, "/auth/google") {
		return "/auth/google"
	}
	return ""
}

func ClienteIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		for _, item := range strings.Split(forwarded, ",") {
			ip := strings.TrimSpace(item)
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
