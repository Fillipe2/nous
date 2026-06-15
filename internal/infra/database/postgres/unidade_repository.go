package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/unidade"
	"nous/internal/infra/database/postgres/internal/scanner"
	"nous/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// UnidadeRepository implementa unidade.Repository para PostgreSQL.
type UnidadeRepository struct {
	connection *Postgres
}

var _ unidade.Repository = (*UnidadeRepository)(nil)

// NewUnidadeRepository constrói o repositório de unidades.
func NewUnidadeRepository(connection *Postgres) (*UnidadeRepository, error) {
	return &UnidadeRepository{connection: connection}, nil
}

func (r *UnidadeRepository) Salvar(ctx context.Context, n *unidade.Unidade) error {
	n.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarUnidade, pgx.NamedArgs{
		"id":            n.Id,
		"negocio_id":    n.NegocioId,
		"nome":          n.Nome,
		"status":        string(n.Status),
		"criado_em":     n.CriadoEm,
		"atualizado_em": n.AtualizadoEm,
		"version":       n.Version,
	})
	if err != nil {
		return fmt.Errorf("salvar unidade: %w", err)
	}
	return nil
}

func (r *UnidadeRepository) BuscarPorID(ctx context.Context, id string) (*unidade.Unidade, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarUnidadePorID, pgx.NamedArgs{"id": id})
	n, err := scanner.Unidade(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, unidade.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar unidade por id: %w", err)
	}
	return n, nil
}

func (r *UnidadeRepository) Atualizar(ctx context.Context, n *unidade.Unidade) error {
	stmts := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, stmts.AtualizarUnidade, pgx.NamedArgs{
		"nome":             n.Nome,
		"status":           string(n.Status),
		"atualizado_em":    n.AtualizadoEm,
		"version":          n.Version,
		"version_anterior": n.Version - 1,
		"id":               n.Id,
	})
	if err != nil {
		return fmt.Errorf("atualizar unidade: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return unidade.ErrVersaoConflito
	}
	return nil
}

func (r *UnidadeRepository) SoftDelete(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SoftDeleteUnidade, pgx.NamedArgs{
		"atualizado_em": time.Now(),
		"id":            id,
	})
	if err != nil {
		return fmt.Errorf("soft delete unidade: %w", err)
	}
	return nil
}

func (r *UnidadeRepository) Listar(ctx context.Context, filtro unidade.FiltroListagem) ([]*unidade.Unidade, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina
	query := `SELECT n.id, n.negocio_id, n.nome, n.status, n.criado_em, n.atualizado_em, n.version FROM unidade n`
	args := []any{}
	argIdx := 1
	if filtro.UsuarioId != "" {
		query += " INNER JOIN usuario_negocio uc ON uc.negocio_id=n.negocio_id"
	}
	query += " WHERE n.status != 'excluido'"
	if filtro.UsuarioId != "" {
		query += fmt.Sprintf(" AND uc.usuario_id=$%d", argIdx)
		args = append(args, filtro.UsuarioId)
		argIdx++
	}
	if filtro.NegocioId != "" {
		query += fmt.Sprintf(" AND n.negocio_id=$%d", argIdx)
		args = append(args, filtro.NegocioId)
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
		return nil, 0, fmt.Errorf("contar unidades: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "n.nome", "status": "n.status", "criado_em": "n.criado_em", "atualizado_em": "n.atualizado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)
	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar unidades: %w", err)
	}
	defer rows.Close()

	var unidades []*unidade.Unidade
	for rows.Next() {
		n, err := scanner.Unidade(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear unidade: %w", err)
		}
		unidades = append(unidades, n)
	}
	return unidades, total, nil
}
