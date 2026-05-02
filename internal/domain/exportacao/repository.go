package exportacao

import "context"

type FiltroListagem struct {
	ClienteId     string
	LandingPageId string
	UsuarioId     string
	NomeArquivo   string
	DataInicio    string
	DataFim       string
	Pagina        int
	PorPagina     int
	OrdenarPor    string
	Ordem         string
}

type Repository interface {
	Salvar(ctx context.Context, e *Exportacao) error
	BuscarPorID(ctx context.Context, id string) (*Exportacao, error)
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Exportacao, int, error)
}
