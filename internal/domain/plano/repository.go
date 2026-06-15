package plano

import "context"

type Repository interface {
	BuscarPlanoAtivoPorNegocio(ctx context.Context, negocioId string) (*Plano, error)
}
