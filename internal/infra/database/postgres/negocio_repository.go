package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/negocio"
	"nous/internal/infra/database/postgres/internal/scanner"
	"nous/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// NegocioRepository implementa negocio.Repository para PostgreSQL.
type NegocioRepository struct {
	connection *Postgres
}

var _ negocio.Repository = (*NegocioRepository)(nil)

// NewNegocioRepository constrói o repositório de negócios.
func NewNegocioRepository(connection *Postgres) (*NegocioRepository, error) {
	return &NegocioRepository{connection: connection}, nil
}

func (r *NegocioRepository) Salvar(ctx context.Context, n *negocio.Negocio) error {
	n.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarNegocio, pgx.NamedArgs{
		"id":            n.Id,
		"cliente_id":    n.ClienteId,
		"nome":          n.Nome,
		"status":        string(n.Status),
		"criado_em":     n.CriadoEm,
		"atualizado_em": n.AtualizadoEm,
		"version":       n.Version,
	})
	if err != nil {
		return fmt.Errorf("salvar negocio: %w", err)
	}
	return nil
}

func (r *NegocioRepository) BuscarPorID(ctx context.Context, id string) (*negocio.Negocio, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarNegocioPorID, pgx.NamedArgs{"id": id})
	n, err := scanner.Negocio(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, negocio.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar negocio por id: %w", err)
	}
	return n, nil
}

func (r *NegocioRepository) Atualizar(ctx context.Context, n *negocio.Negocio) error {
	stmts := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, stmts.AtualizarNegocio, pgx.NamedArgs{
		"nome":             n.Nome,
		"status":           string(n.Status),
		"atualizado_em":    n.AtualizadoEm,
		"version":          n.Version,
		"version_anterior": n.Version - 1,
		"id":               n.Id,
	})
	if err != nil {
		return fmt.Errorf("atualizar negocio: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return negocio.ErrVersaoConflito
	}
	return nil
}

func (r *NegocioRepository) SoftDelete(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SoftDeleteNegocio, pgx.NamedArgs{
		"atualizado_em": time.Now(),
		"id":            id,
	})
	if err != nil {
		return fmt.Errorf("soft delete negocio: %w", err)
	}
	return nil
}

func (r *NegocioRepository) Listar(ctx context.Context, filtro negocio.FiltroListagem) ([]*negocio.Negocio, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina
	query := `SELECT n.id, n.cliente_id, n.nome, n.status, n.criado_em, n.atualizado_em, n.version FROM negocio n`
	args := []any{}
	argIdx := 1
	if filtro.UsuarioId != "" {
		query += " INNER JOIN usuario_cliente uc ON uc.cliente_id=n.cliente_id"
	}
	query += " WHERE n.status != 'excluido'"
	if filtro.UsuarioId != "" {
		query += fmt.Sprintf(" AND uc.usuario_id=$%d", argIdx)
		args = append(args, filtro.UsuarioId)
		argIdx++
	}
	if filtro.ClienteId != "" {
		query += fmt.Sprintf(" AND n.cliente_id=$%d", argIdx)
		args = append(args, filtro.ClienteId)
		argIdx++
	}
	if filtro.Nome != "" {
		query += fmt.Sprintf(" AND n.nome ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Nome+"%")
		argIdx++
	}
	if filtro.Status != "" {
		query += fmt.Sprintf(" AND n.status=$%d", argIdx)
		args = append(args, filtro.Status)
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND n.criado_em >= $%d", argIdx)
		args = append(args, dataInicioFiltro(filtro.DataInicio))
		argIdx++
	}
	if filtro.DataFim != "" {
		operador, valor := dataFimFiltro(filtro.DataFim)
		query += fmt.Sprintf(" AND n.criado_em %s $%d", operador, argIdx)
		args = append(args, valor)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar negocios: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "n.nome", "status": "n.status", "criado_em": "n.criado_em", "atualizado_em": "n.atualizado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)
	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar negocios: %w", err)
	}
	defer rows.Close()

	var negocios []*negocio.Negocio
	for rows.Next() {
		n, err := scanner.Negocio(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear negocio: %w", err)
		}
		negocios = append(negocios, n)
	}
	return negocios, total, nil
}
