package dashboard

import "context"

type Repository interface {
	ObterResumoNegocio(ctx context.Context, negocioId string, filtro FiltroPeriodo) (*ResumoNegocio, error)
	ObterResumoUsuario(ctx context.Context, usuarioId string, filtro FiltroPeriodo) (*ResumoUsuario, error)
	ObterResumoAdmin(ctx context.Context, filtro FiltroPeriodo) (*ResumoAdmin, error)
}
