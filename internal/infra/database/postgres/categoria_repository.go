package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/categoria"
	"nous/internal/infra/database/postgres/internal/scanner"
	"nous/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// CategoriaRepository implementa categoria.Repository para PostgreSQL.
type CategoriaRepository struct {
	connection *Postgres
}

var _ categoria.Repository = (*CategoriaRepository)(nil)

// NewCategoriaRepository constrói o repositório de categorias.
func NewCategoriaRepository(connection *Postgres) (*CategoriaRepository, error) {
	return &CategoriaRepository{connection: connection}, nil
}

func (r *CategoriaRepository) Salvar(ctx context.Context, c *categoria.Categoria) error {
	c.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarCategoria, pgx.NamedArgs{
		"id":            c.Id,
		"nome":          c.Nome,
		"status":        string(c.Status),
		"criado_em":     c.CriadoEm,
		"atualizado_em": c.AtualizadoEm,
		"version":       c.Version,
	})
	if err != nil {
		return fmt.Errorf("salvar categoria: %w", err)
	}
	return nil
}

func (r *CategoriaRepository) BuscarPorID(ctx context.Context, id string) (*categoria.Categoria, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarCategoriaPorID, pgx.NamedArgs{"id": id})
	c, err := scanner.Categoria(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, categoria.ErrNaoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("buscar categoria por id: %w", err)
	}
	return c, nil
}

func (r *CategoriaRepository) Atualizar(ctx context.Context, c *categoria.Categoria) error {
	stmts := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, stmts.AtualizarCategoria, pgx.NamedArgs{
		"nome":             c.Nome,
		"status":           string(c.Status),
		"atualizado_em":    c.AtualizadoEm,
		"version":          c.Version,
		"version_anterior": c.Version - 1,
		"id":               c.Id,
	})
	if err != nil {
		return fmt.Errorf("atualizar categoria: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return categoria.ErrVersaoConflito
	}
	return nil
}

func (r *CategoriaRepository) SoftDelete(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SoftDeleteCategoria, pgx.NamedArgs{
		"atualizado_em": time.Now(),
		"id":            id,
	})
	if err != nil {
		return fmt.Errorf("soft delete categoria: %w", err)
	}
	return nil
}

func (r *CategoriaRepository) Listar(ctx context.Context, filtro categoria.FiltroListagem) ([]*categoria.Categoria, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina
	query := `SELECT id, nome, status, criado_em, atualizado_em, version FROM categoria WHERE status != 'excluido'`
	args := []any{}
	argIdx := 1
	if filtro.Nome != "" {
		query += fmt.Sprintf(" AND nome ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Nome+"%")
		argIdx++
	}
	if filtro.Status != "" {
		query += fmt.Sprintf(" AND status=$%d", argIdx)
		args = append(args, filtro.Status)
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND criado_em >= $%d", argIdx)
		args = append(args, dataInicioFiltro(filtro.DataInicio))
		argIdx++
	}
	if filtro.DataFim != "" {
		operador, valor := dataFimFiltro(filtro.DataFim)
		query += fmt.Sprintf(" AND criado_em %s $%d", operador, argIdx)
		args = append(args, valor)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar categorias: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "nome", map[string]string{
		"nome": "nome", "status": "status", "criado_em": "criado_em", "atualizado_em": "atualizado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)
	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar categorias: %w", err)
	}
	defer rows.Close()

	var categorias []*categoria.Categoria
	for rows.Next() {
		c, err := scanner.Categoria(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear categoria: %w", err)
		}
		categorias = append(categorias, c)
	}
	return categorias, total, nil
}
