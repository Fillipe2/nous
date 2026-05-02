package plano

import "context"

type Repository interface {
	BuscarPlanoAtivoPorCliente(ctx context.Context, clienteId string) (*Plano, error)
}
