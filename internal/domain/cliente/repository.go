package cliente

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
	Salvar(ctx context.Context, c *Cliente) error
	SalvarComVinculo(ctx context.Context, c *Cliente, uc *UsuarioCliente) error
	BuscarPorID(ctx context.Context, id string) (*Cliente, error)
	Atualizar(ctx context.Context, c *Cliente) error
	SoftDelete(ctx context.Context, id string) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Cliente, int, error)
	VincularUsuario(ctx context.Context, uc *UsuarioCliente) error
	ListarPorUsuario(ctx context.Context, usuarioId string) ([]*Cliente, error)
	VerificarAcessoUsuario(ctx context.Context, clienteId, usuarioId string) error
}
