package auditoria

import "context"

type FiltroListagem struct {
	UsuarioId  string
	Acao       string
	Entidade   string
	EntidadeId string
	StatusCode string
	RequestId  string
	DataInicio string
	DataFim    string
	Pagina     int
	PorPagina  int
	OrdenarPor string
	Ordem      string
}

type Repository interface {
	Salvar(ctx context.Context, a *Auditoria) error
	BuscarPorID(ctx context.Context, id string) (*Auditoria, error)
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Auditoria, int, error)
}
