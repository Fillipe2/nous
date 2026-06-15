package usuario

import "unicode/utf8"

const (
	// SenhaTamanhoMinimo segue a recomendação NIST SP 800-63B (mínimo de 8).
	SenhaTamanhoMinimo = 8
	// SenhaTamanhoMaximo respeita o limite de 72 bytes do bcrypt; bytes além
	// disso são silenciosamente ignorados na geração do hash.
	SenhaTamanhoMaximo = 72
)

// ValidarSenha aplica a política mínima de senha do sistema.
//
// O comprimento mínimo é medido em caracteres (runes) para ser justo com
// senhas multibyte; o máximo é medido em bytes por causa do limite do bcrypt.
func ValidarSenha(senha string) error {
	if utf8.RuneCountInString(senha) < SenhaTamanhoMinimo || len(senha) > SenhaTamanhoMaximo {
		return ErrSenhaFraca
	}
	return nil
}
