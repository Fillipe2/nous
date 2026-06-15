package negocio

import (
	"errors"
	"testing"
)

func TestNormalizarDocumento(t *testing.T) {
	casos := map[string]string{
		"529.982.247-25":     "52998224725",
		"12.ABC.345/01DE-35": "12ABC34501DE35",
		" 12abc34501de35 ":   "12ABC34501DE35",
		"11.222.333/0001-81": "11222333000181",
		"":                   "",
	}
	for entrada, esperado := range casos {
		if got := NormalizarDocumento(entrada); got != esperado {
			t.Errorf("NormalizarDocumento(%q) = %q, esperado %q", entrada, got, esperado)
		}
	}
}

func TestValidarDocumentoValidos(t *testing.T) {
	validos := []string{
		"52998224725",    // CPF
		"00000000191",    // CPF com zeros à esquerda
		"11222333000181", // CNPJ numérico
		"00001111000112", // CNPJ com zeros à esquerda
		"12ABC34501DE35", // CNPJ alfanumérico (exemplo oficial Serpro)
	}
	for _, doc := range validos {
		if err := ValidarDocumento(doc); err != nil {
			t.Errorf("ValidarDocumento(%q) = %v, esperava válido", doc, err)
		}
	}
}

func TestValidarDocumentoInvalidos(t *testing.T) {
	invalidos := []string{
		"",               // vazio
		"123",            // tamanho inválido
		"52998224724",    // CPF com DV errado
		"11111111111",    // CPF com todos os dígitos iguais
		"11222333000180", // CNPJ com DV errado
		"11111111111111", // CNPJ com todos os caracteres iguais
		"12ABC34501DE34", // CNPJ alfanumérico com DV errado
		"12ABC34501DEA5", // DV não numérico
		"1229982247253",  // 13 caracteres (nem CPF nem CNPJ)
	}
	for _, doc := range invalidos {
		if err := ValidarDocumento(doc); !errors.Is(err, ErrDocumentoInvalido) {
			t.Errorf("ValidarDocumento(%q) = %v, esperava ErrDocumentoInvalido", doc, err)
		}
	}
}
