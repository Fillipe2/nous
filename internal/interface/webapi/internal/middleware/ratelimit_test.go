package middleware

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestMemoryRateLimiterRespeitaLimite(t *testing.T) {
	resetRateLimiterStateForTest()
	store := memoryRateLimiter{}

	for tentativa := 1; tentativa <= 2; tentativa++ {
		excedeu, err := store.Excedeu(context.Background(), "teste", 2, time.Minute)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if excedeu {
			t.Fatalf("tentativa %d nao deveria exceder limite", tentativa)
		}
	}

	excedeu, err := store.Excedeu(context.Background(), "teste", 2, time.Minute)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !excedeu {
		t.Fatal("terceira tentativa deveria exceder limite")
	}
}

func TestMemoryRateLimiterResetaJanelaExpirada(t *testing.T) {
	resetRateLimiterStateForTest()
	store := memoryRateLimiter{}

	if excedeu, err := store.Excedeu(context.Background(), "teste", 1, time.Minute); err != nil || excedeu {
		t.Fatalf("primeira tentativa inesperada: excedeu=%v err=%v", excedeu, err)
	}

	rateState.Lock()
	bucket := rateState.buckets["teste"]
	bucket.resetEm = time.Now().Add(-time.Second)
	rateState.buckets["teste"] = bucket
	rateState.Unlock()

	if excedeu, err := store.Excedeu(context.Background(), "teste", 1, time.Minute); err != nil || excedeu {
		t.Fatalf("janela expirada deveria reiniciar limite: excedeu=%v err=%v", excedeu, err)
	}
}

func TestClienteIPIgnoraForwardedPorPadrao(t *testing.T) {
	// Sem TRUSTED_PROXY_COUNT, o XFF (forjável) deve ser ignorado.
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("criar request: %v", err)
	}
	request.RemoteAddr = "10.0.0.10:12345"
	request.Header.Set("X-Forwarded-For", "203.0.113.10")

	if ip := ClienteIP(request); ip != "10.0.0.10" {
		t.Fatalf("ip esperado do RemoteAddr (XFF ignorado), recebido %s", ip)
	}
}

func TestClienteIPUsaForwardedComProxyConfiavel(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_COUNT", "1")
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("criar request: %v", err)
	}
	// RemoteAddr é o proxy confiável; o cliente real é o último IP do XFF.
	request.RemoteAddr = "10.0.0.1:443"
	request.Header.Set("X-Forwarded-For", "203.0.113.10")

	if ip := ClienteIP(request); ip != "203.0.113.10" {
		t.Fatalf("ip esperado do XFF via proxy confiavel, recebido %s", ip)
	}
}

func TestRateLimiterUsaRedisURLConfiguradaSemRelerEnv(t *testing.T) {
	resetRateLimiterStateForTest()
	ConfigurarRateLimiter("redis://localhost:6379/0")
	t.Setenv("REDIS_URL", "")

	if redisURL := redisURLConfigurada(); redisURL != "redis://localhost:6379/0" {
		t.Fatalf("redis url configurada deveria ser mantida, recebida %q", redisURL)
	}
}

func resetRateLimiterStateForTest() {
	rateState.Lock()
	rateState.buckets = make(map[string]rateBucket)
	rateState.Unlock()
	ConfigurarRateLimiter("")
	CloseRateLimiter()
}
