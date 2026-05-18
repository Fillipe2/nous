package faturamento

import (
	"context"
	"time"
)

// UsoRepository persiste e consulta eventos de uso por cliente.
type UsoRepository interface {
	Registrar(ctx context.Context, uso *UsoClienteEvento) error
}

// FaturamentoRepository gerencia o faturamento mensal por cliente.
type FaturamentoRepository interface {
	ObterOuCriarCompetencia(ctx context.Context, clienteId string, competencia time.Time) (*FaturamentoMensal, error)
	AdicionarUso(ctx context.Context, clienteId string, competencia time.Time, centavos int) error
	FecharCompetencia(ctx context.Context, clienteId string, competencia time.Time) error
}

// BiRepository persiste métricas diárias de BI.
type BiRepository interface {
	IncrementarMetrica(ctx context.Context, clienteId string, data time.Time, metrica string, valor float64, dimensoes map[string]string) error
}
