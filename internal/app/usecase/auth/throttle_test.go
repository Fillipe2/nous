package auth

import (
	"testing"
	"time"
)

func TestLoginThrottleBloqueiaAposLimite(t *testing.T) {
	th := novoLoginThrottle(3, time.Minute)
	const email = "User@Sognix.Test"

	for i := range 3 {
		if th.Bloqueado(email) {
			t.Fatalf("nao deveria bloquear antes do limite (i=%d)", i)
		}
		th.RegistrarFalha(email)
	}

	if !th.Bloqueado(email) {
		t.Fatal("deveria bloquear apos atingir o limite")
	}
	// A chave deve ser normalizada (case-insensitive, sem espacos).
	if !th.Bloqueado("  user@sognix.test  ") {
		t.Fatal("bloqueio deveria ser case-insensitive e ignorar espacos")
	}
}

func TestLoginThrottleSucessoZeraContador(t *testing.T) {
	th := novoLoginThrottle(3, time.Minute)
	const email = "user@sognix.test"

	th.RegistrarFalha(email)
	th.RegistrarFalha(email)
	th.RegistrarSucesso(email)
	th.RegistrarFalha(email)

	if th.Bloqueado(email) {
		t.Fatal("login bem-sucedido deveria ter zerado o contador")
	}
}

func TestLoginThrottleJanelaExpira(t *testing.T) {
	th := novoLoginThrottle(1, 10*time.Millisecond)
	const email = "user@sognix.test"

	th.RegistrarFalha(email)
	if !th.Bloqueado(email) {
		t.Fatal("deveria bloquear imediatamente com limite 1")
	}

	time.Sleep(20 * time.Millisecond)
	if th.Bloqueado(email) {
		t.Fatal("bloqueio deveria expirar apos a janela")
	}
}
