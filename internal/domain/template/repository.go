package template

import "context"

type FiltroListagem struct {
	ClienteId                string
	CategoriaId              string
	Tipo                     string
	Origem                   string
	Nome                     string
	Status                   string
	DataInicio               string
	DataFim                  string
	PermitirBaseCustomizavel bool
	Pagina                   int
	PorPagina                int
	OrdenarPor               string
	Ordem                    string
}

type Repository interface {
	Salvar(ctx context.Context, t *Template) error
	BuscarPorID(ctx context.Context, id string) (*Template, error)
	SoftDelete(ctx context.Context, id string) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Template, int, error)
}
