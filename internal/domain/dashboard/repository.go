package dashboard

import "context"

type Repository interface {
	ObterResumoCliente(ctx context.Context, clienteId string, filtro FiltroPeriodo) (*ResumoCliente, error)
	ObterResumoUsuario(ctx context.Context, usuarioId string, filtro FiltroPeriodo) (*ResumoUsuario, error)
	ObterResumoAdmin(ctx context.Context, filtro FiltroPeriodo) (*ResumoAdmin, error)
}
