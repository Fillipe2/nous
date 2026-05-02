package postgres

import "strings"

func colunaOrdenacao(valor, padrao string, permitidas map[string]string) string {
	valor = strings.ToLower(strings.TrimSpace(valor))
	if coluna, ok := permitidas[valor]; ok {
		return coluna
	}
	return permitidas[padrao]
}

func direcaoOrdenacao(valor string) string {
	if strings.EqualFold(strings.TrimSpace(valor), "asc") {
		return "ASC"
	}
	return "DESC"
}
