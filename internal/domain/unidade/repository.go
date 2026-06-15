package unidade

import "context"

type FiltroListagem struct {
	NegocioId  string
	UsuarioId  string
	Nome       string
	Status     string
	DataInicio string
	DataFim    string
	Pagina     int
	PorPagina  int
	OrdenarPor string
	Ordem      string
}

type Repository interface {
	Salvar(ctx context.Context, n *Unidade) error
	BuscarPorID(ctx context.Context, id string) (*Unidade, error)
	Atualizar(ctx context.Context, n *Unidade) error
	SoftDelete(ctx context.Context, id string) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Unidade, int, error)
}
