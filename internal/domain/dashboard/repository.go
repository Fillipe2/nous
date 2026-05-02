package dashboard

import "context"

type Repository interface {
	ObterResumoCliente(ctx context.Context, clienteId string, filtro FiltroPeriodo) (*ResumoCliente, error)
	ObterResumoAdmin(ctx context.Context, filtro FiltroPeriodo) (*ResumoAdmin, error)
}
