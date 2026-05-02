package negocio

import "context"

type FiltroListagem struct {
	ClienteId  string
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
	Salvar(ctx context.Context, n *Negocio) error
	BuscarPorID(ctx context.Context, id string) (*Negocio, error)
	Atualizar(ctx context.Context, n *Negocio) error
	SoftDelete(ctx context.Context, id string) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Negocio, int, error)
}
