package postgres

import (
	"testing"
	"time"
)

func TestDataFimFiltroDataApenasIncluiDiaInteiro(t *testing.T) {
	operador, valor := dataFimFiltro("2026-05-02")
	if operador != "<" {
		t.Fatalf("operador esperado '<', recebido %q", operador)
	}
	data, ok := valor.(time.Time)
	if !ok {
		t.Fatalf("valor esperado time.Time, recebido %T", valor)
	}
	esperado := time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)
	if !data.Equal(esperado) {
		t.Fatalf("data final esperada %s, recebida %s", esperado, data)
	}
}
