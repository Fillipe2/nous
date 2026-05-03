package lead

import (
	"context"
	"time"
)

type FiltroListagem struct {
	ClienteId     string
	LandingPageId string
	Nome          string
	Email         string
	Status        string
	UTMSource     string
	UTMCampaign   string
	DataInicio    string
	DataFim       string
	Pagina        int
	PorPagina     int
	OrdenarPor    string
	Ordem         string
}

type AtualizacaoStatus struct {
	Status Status
	Motivo string
}

type Repository interface {
	Salvar(ctx context.Context, l *Lead) error
	BuscarPorID(ctx context.Context, id string) (*Lead, error)
	BuscarDuplicado(ctx context.Context, email, landingPageId string, desde time.Time) (*Lead, error)
	AtualizarStatus(ctx context.Context, id string, atualizacao AtualizacaoStatus, agora time.Time) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Lead, int, error)
	Anonimizar(ctx context.Context, id string, agora time.Time) error
	AnonimizarExpirados(ctx context.Context, antesDe time.Time) (int, error)
}
