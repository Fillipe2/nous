package usuario

import (
	"errors"
	"strings"
	"testing"
)

func TestValidarSenha(t *testing.T) {
	casos := []struct {
		nome    string
		senha   string
		querErr bool
	}{
		{"vazia", "", true},
		{"curta", "1234567", true},
		{"minima", "12345678", false},
		{"forte", "uma-senha-bem-segura", false},
		{"multibyte_no_limite", strings.Repeat("á", 8), false},
		{"acima_do_maximo_bcrypt", strings.Repeat("a", 73), true},
		{"no_maximo_bcrypt", strings.Repeat("a", 72), false},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			err := ValidarSenha(c.senha)
			if c.querErr && !errors.Is(err, ErrSenhaFraca) {
				t.Fatalf("esperava ErrSenhaFraca, recebeu %v", err)
			}
			if !c.querErr && err != nil {
				t.Fatalf("nao esperava erro, recebeu %v", err)
			}
		})
	}
}
