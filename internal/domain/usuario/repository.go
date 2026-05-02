package usuario

import (
	"context"
	"time"
)

type FiltroListagem struct {
	Id         string
	ClienteId  string
	Nome       string
	Email      string
	Role       string
	Status     string
	DataInicio string
	DataFim    string
	Pagina     int
	PorPagina  int
	OrdenarPor string
	Ordem      string
}

type Repository interface {
	Salvar(ctx context.Context, u *Usuario) error
	BuscarPorID(ctx context.Context, id string) (*Usuario, error)
	BuscarPorEmail(ctx context.Context, email string) (*Usuario, error)
	Atualizar(ctx context.Context, u *Usuario) error
	SoftDelete(ctx context.Context, id string) error
	Listar(ctx context.Context, filtro FiltroListagem) ([]*Usuario, int, error)
	SalvarAuth(ctx context.Context, auth *UsuarioAuth) error
	BuscarAuthPorProvider(ctx context.Context, tipo, providerID string) (*UsuarioAuth, error)
	SalvarRefreshToken(ctx context.Context, token *RefreshToken) error
	BuscarRefreshTokenPorHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevogarRefreshToken(ctx context.Context, tokenHash string, revogadoEm time.Time) error
	ExpurgarRefreshTokens(ctx context.Context, antesDe time.Time) error
}
