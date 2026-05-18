package lead_aquisicao

import (
	"context"
	"time"
)

type FiltroListagem struct {
	Nome        string
	Email       string
	StatusFunil string
	Porte       string
	Cidade      string
	Estado      string
	UTMCampaign string
	DataInicio  string
	DataFim     string
	Pagina      int
	PorPagina   int
	OrdenarPor  string
	Ordem       string
}

type AtualizacaoFunil struct {
	StatusFunil           StatusFunil
	ConvertidoEmUsuarioId string
}

type Repository interface {
	Salvar(ctx context.Context, la *LeadAquisicao) error
	BuscarPorID(ctx context.Context, id string) (*LeadAquisicao, error)
	BuscarPorEmail(ctx context.Context, email string) (*LeadAquisicao, error)
	AtualizarFunil(ctx context.Context, id string, atualizacao AtualizacaoFunil, agora time.Time) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*LeadAquisicao, int, error)
}
