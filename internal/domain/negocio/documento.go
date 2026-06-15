package negocio

import "strings"

// O documento do Negócio é CPF (profissional autônomo) ou CNPJ (empresa).
//
// É armazenado apenas com os caracteres significativos, em maiúsculas e sem
// separadores, preservando zeros à esquerda — por isso é sempre string, nunca
// número.
//
// CPF: 11 dígitos numéricos.
// CNPJ: 14 posições, sendo as 12 primeiras alfanuméricas ([0-9A-Z]) e as 2
// últimas (dígitos verificadores) numéricas. Esse é o formato oficial da Receita
// Federal a partir de 2026 e é retrocompatível com o CNPJ totalmente numérico,
// pois o DV usa módulo 11 sobre (ASCII do caractere - 48).

// NormalizarDocumento remove separadores comuns e coloca em maiúsculas,
// preservando zeros à esquerda e letras. Não valida.
func NormalizarDocumento(doc string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(doc)) {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ValidarDocumento valida um documento já normalizado (ver NormalizarDocumento):
// CPF de 11 dígitos ou CNPJ de 14 posições (incl. alfanumérico).
func ValidarDocumento(doc string) error {
	switch len(doc) {
	case 11:
		if cpfValido(doc) {
			return nil
		}
	case 14:
		if cnpjValido(doc) {
			return nil
		}
	}
	return ErrDocumentoInvalido
}

func cpfValido(cpf string) bool {
	for i := range 11 {
		if cpf[i] < '0' || cpf[i] > '9' {
			return false
		}
	}
	if caracteresIguais(cpf) {
		return false
	}
	return int(cpf[9]-'0') == digitoCPF(cpf, 9) && int(cpf[10]-'0') == digitoCPF(cpf, 10)
}

// digitoCPF calcula um dígito verificador de CPF sobre os primeiros `ate` dígitos.
func digitoCPF(cpf string, ate int) int {
	soma := 0
	peso := ate + 1
	for i := range ate {
		soma += int(cpf[i]-'0') * peso
		peso--
	}
	resto := (soma * 10) % 11
	if resto == 10 {
		return 0
	}
	return resto
}

func cnpjValido(cnpj string) bool {
	for i := range 12 {
		c := cnpj[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	// Dígitos verificadores são sempre numéricos.
	if cnpj[12] < '0' || cnpj[12] > '9' || cnpj[13] < '0' || cnpj[13] > '9' {
		return false
	}
	if caracteresIguais(cnpj[:12]) {
		return false
	}
	dv1 := digitoCNPJ(cnpj[:12])
	dv2 := digitoCNPJ(cnpj[:12] + string(rune('0'+dv1)))
	return int(cnpj[12]-'0') == dv1 && int(cnpj[13]-'0') == dv2
}

// digitoCNPJ calcula um dígito verificador de CNPJ (módulo 11) sobre `base`,
// usando pesos cíclicos 2..9 da direita para a esquerda e o valor alfanumérico
// (ASCII - 48) de cada caractere.
func digitoCNPJ(base string) int {
	soma := 0
	peso := 2
	for i := len(base) - 1; i >= 0; i-- {
		soma += (int(base[i]) - 48) * peso
		peso++
		if peso > 9 {
			peso = 2
		}
	}
	resto := soma % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}

func caracteresIguais(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return len(s) > 0
}
