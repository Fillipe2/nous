package landingpage

import (
	"context"
	"time"
)

type FiltroListagem struct {
	NegocioId   string
	UnidadeId   string
	CategoriaId string
	TemplateId  string
	Tipo        string
	Status      string
	Nome        string
	IsDraft     string
	DataInicio  string
	DataFim     string
	UsuarioId   string
	Pagina      int
	PorPagina   int
	OrdenarPor  string
	Ordem       string
}

type Repository interface {
	Salvar(ctx context.Context, lp *LandingPage) error
	BuscarPorID(ctx context.Context, id string) (*LandingPage, error)
	Atualizar(ctx context.Context, lp *LandingPage) error
	SoftDelete(ctx context.Context, id string) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*LandingPage, int, error)
	ContarPorNegocio(ctx context.Context, negocioId string) (int, error)
	AtualizarAutosave(ctx context.Context, id string, dados []byte, version int) (int, error)
	AtualizarLock(ctx context.Context, id string, usuarioId *string) error
	VerificarAcessoUsuario(ctx context.Context, landingPageId, usuarioId string) error
	AtualizarPublicacao(ctx context.Context, id, urlPublica, deployId string, publicadaEm time.Time) error
	AtualizarDespublicacao(ctx context.Context, id string) error
}
