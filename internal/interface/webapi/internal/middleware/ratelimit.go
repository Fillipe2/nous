package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	limiteGlobalPorMinuto   = 300
	limiteSensivelPorMinuto = 20
)

type rateBucket struct {
	contador int
	resetEm  time.Time
}

var rateState = struct {
	sync.Mutex
	buckets map[string]rateBucket
}{buckets: make(map[string]rateBucket)}

func LimitarTaxa(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ipCliente(r)
		classe := classeRota(r)
		if excedeuLimite("global:"+ip, limiteGlobalPorMinuto) || (classe != "" && excedeuLimite("sensivel:"+ip+":"+classe, limiteSensivelPorMinuto)) {
			w.Header().Set("Retry-After", "60")
			responderErro(w, r, http.StatusTooManyRequests, "muitas requisicoes, tente novamente em instantes")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func excedeuLimite(chave string, limite int) bool {
	agora := time.Now()
	rateState.Lock()
	defer rateState.Unlock()

	bucket := rateState.buckets[chave]
	if bucket.resetEm.IsZero() || agora.After(bucket.resetEm) {
		rateState.buckets[chave] = rateBucket{contador: 1, resetEm: agora.Add(time.Minute)}
		limparBucketsExpirados(agora)
		return false
	}

	bucket.contador++
	rateState.buckets[chave] = bucket
	return bucket.contador > limite
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
	if r.Method == http.MethodPost && (path == "/auth/login" || path == "/auth/refresh" || path == "/usuarios" || path == "/leads") {
		return path
	}
	if strings.HasPrefix(path, "/auth/google") {
		return "/auth/google"
	}
	return ""
}

func ipCliente(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
