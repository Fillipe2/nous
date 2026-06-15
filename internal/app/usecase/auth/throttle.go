package auth

import (
	"strings"
	"sync"
	"time"
)

const (
	// maxTentativasLogin é o número de falhas de login por conta dentro da
	// janela antes do bloqueio temporário.
	maxTentativasLogin = 10
	// janelaTentativasLogin é a duração da janela de contagem; após ela, o
	// contador da conta zera automaticamente (bloqueio sempre temporário,
	// nunca permanente, para evitar DoS de travamento de conta).
	janelaTentativasLogin = 15 * time.Minute
)

type tentativaLogin struct {
	contador int
	resetEm  time.Time
}

// loginThrottle limita tentativas de login malsucedidas por conta (e-mail),
// complementando o rate limit por IP do middleware. É um controle best-effort
// em memória, por instância; em deploy multi-instância cada réplica mantém o
// próprio contador.
type loginThrottle struct {
	mu         sync.Mutex
	tentativas map[string]tentativaLogin
	max        int
	janela     time.Duration
}

func novoLoginThrottle(max int, janela time.Duration) *loginThrottle {
	return &loginThrottle{
		tentativas: make(map[string]tentativaLogin),
		max:        max,
		janela:     janela,
	}
}

func chaveLogin(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Bloqueado informa se a conta atingiu o limite de falhas na janela atual.
func (t *loginThrottle) Bloqueado(email string) bool {
	chave := chaveLogin(email)
	t.mu.Lock()
	defer t.mu.Unlock()

	tent, ok := t.tentativas[chave]
	if !ok {
		return false
	}
	if time.Now().After(tent.resetEm) {
		delete(t.tentativas, chave)
		return false
	}
	return tent.contador >= t.max
}

// RegistrarFalha incrementa o contador de falhas da conta na janela atual.
func (t *loginThrottle) RegistrarFalha(email string) {
	chave := chaveLogin(email)
	agora := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.limparExpirados(agora)

	tent := t.tentativas[chave]
	if tent.resetEm.IsZero() || agora.After(tent.resetEm) {
		t.tentativas[chave] = tentativaLogin{contador: 1, resetEm: agora.Add(t.janela)}
		return
	}
	tent.contador++
	t.tentativas[chave] = tent
}

// RegistrarSucesso zera o contador da conta após um login bem-sucedido.
func (t *loginThrottle) RegistrarSucesso(email string) {
	chave := chaveLogin(email)
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tentativas, chave)
}

// limparExpirados remove entradas cuja janela já passou, mantendo o mapa
// limitado às contas com falhas recentes. Deve ser chamado com o mutex preso.
func (t *loginThrottle) limparExpirados(agora time.Time) {
	for chave, tent := range t.tentativas {
		if agora.After(tent.resetEm) {
			delete(t.tentativas, chave)
		}
	}
}
