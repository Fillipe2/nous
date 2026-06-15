package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/plano"
	"nous/internal/infra/database/postgres/internal/statement"
)

type PlanoRepository struct {
	connection *Postgres
}

var _ plano.Repository = (*PlanoRepository)(nil)

func NewPlanoRepository(connection *Postgres) (*PlanoRepository, error) {
	return &PlanoRepository{connection: connection}, nil
}

func (r *PlanoRepository) BuscarPlanoAtivoPorNegocio(ctx context.Context, negocioId string) (*plano.Plano, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarPlanoAtivoPorNegocio, pgx.NamedArgs{"negocio_id": negocioId})

	var (
		p            plano.Plano
		descricao    *string
		limitesBytes []byte
	)
	if err := row.Scan(&p.Id, &p.Nome, &descricao, &p.ValorMensalCentavos, &limitesBytes, &p.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, plano.ErrAssinaturaAtivaNaoEncontrada
		}
		return nil, fmt.Errorf("buscar plano ativo por negocio: %w", err)
	}
	if descricao != nil {
		p.Descricao = *descricao
	}
	if len(limitesBytes) > 0 {
		if err := json.Unmarshal(limitesBytes, &p.Limites); err != nil {
			return nil, fmt.Errorf("parsear limites do plano: %w", err)
		}
	}
	return &p, nil
}
