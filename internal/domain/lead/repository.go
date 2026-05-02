package lead

import "context"

type FiltroListagem struct {
	ClienteId     string
	LandingPageId string
	Nome          string
	Email         string
	Status        string
	DataInicio    string
	DataFim       string
	Pagina        int
	PorPagina     int
	OrdenarPor    string
	Ordem         string
}

type Repository interface {
	Salvar(ctx context.Context, l *Lead) error
	BuscarPorID(ctx context.Context, id string) (*Lead, error)
	AtualizarStatus(ctx context.Context, id string, status Status) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Lead, int, error)
}
