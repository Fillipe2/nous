package negocio

import "context"

type FiltroListagem struct {
	UsuarioId  string
	Nome       string
	Email      string
	Status     string
	DataInicio string
	DataFim    string
	Pagina     int
	PorPagina  int
	OrdenarPor string
	Ordem      string
}

type Repository interface {
	Salvar(ctx context.Context, c *Negocio) error
	SalvarComVinculo(ctx context.Context, c *Negocio, uc *UsuarioNegocio) error
	BuscarPorID(ctx context.Context, id string) (*Negocio, error)
	Atualizar(ctx context.Context, c *Negocio) error
	SoftDelete(ctx context.Context, id string) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Negocio, int, error)
	VincularUsuario(ctx context.Context, uc *UsuarioNegocio) error
	ListarPorUsuario(ctx context.Context, usuarioId string) ([]*Negocio, error)
	VerificarAcessoUsuario(ctx context.Context, negocioId, usuarioId string) error
}
