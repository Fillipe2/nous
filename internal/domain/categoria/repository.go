package categoria

import "context"

type FiltroListagem struct {
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
	Salvar(ctx context.Context, c *Categoria) error
	BuscarPorID(ctx context.Context, id string) (*Categoria, error)
	Atualizar(ctx context.Context, c *Categoria) error
	SoftDelete(ctx context.Context, id string) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Categoria, int, error)
}
