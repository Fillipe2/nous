package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/template"
	"buildnix-backend/internal/infra/database/postgres/internal/scanner"
	"buildnix-backend/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// TemplateRepository implementa template.Repository para PostgreSQL.
type TemplateRepository struct {
	connection *Postgres
}

var _ template.Repository = (*TemplateRepository)(nil)

// NewTemplateRepository constrói o repositório de templates.
func NewTemplateRepository(connection *Postgres) (*TemplateRepository, error) {
	return &TemplateRepository{connection: connection}, nil
}

func (r *TemplateRepository) Salvar(ctx context.Context, t *template.Template) error {
	t.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarTemplate, pgx.NamedArgs{
		"id":            t.Id,
		"cliente_id":    nilableStr(t.ClienteId),
		"categoria_id":  nilableStr(t.CategoriaId),
		"tipo":          string(t.Tipo),
		"origem":        string(t.Origem),
		"nome":          t.Nome,
		"arquivo_url":   nilableStr(t.ArquivoURL),
		"conteudo":      t.Conteudo,
		"status":        string(t.Status),
		"criado_em":     t.CriadoEm,
		"atualizado_em": t.AtualizadoEm,
		"version":       t.Version,
	})
	if err != nil {
		return fmt.Errorf("salvar template: %w", err)
	}
	return nil
}

func (r *TemplateRepository) BuscarPorID(ctx context.Context, id string) (*template.Template, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarTemplatePorID, pgx.NamedArgs{"id": id})
	t, err := scanner.Template(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, template.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar template por id: %w", err)
	}
	return t, nil
}

func (r *TemplateRepository) SoftDelete(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SoftDeleteTemplate, pgx.NamedArgs{
		"atualizado_em": time.Now(),
		"id":            id,
	})
	if err != nil {
		return fmt.Errorf("soft delete template: %w", err)
	}
	return nil
}

func (r *TemplateRepository) Listar(ctx context.Context, filtro template.FiltroListagem) ([]*template.Template, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina
	query := `SELECT id, cliente_id, categoria_id, tipo, origem, nome, arquivo_url, conteudo, status, criado_em, atualizado_em, version FROM template WHERE status != 'excluido'`
	args := []any{}
	argIdx := 1

	if filtro.ClienteId != "" {
		query += fmt.Sprintf(" AND (origem='sognix' OR cliente_id=$%d)", argIdx)
		args = append(args, filtro.ClienteId)
		argIdx++
	}
	if filtro.CategoriaId != "" {
		query += fmt.Sprintf(" AND categoria_id=$%d", argIdx)
		args = append(args, filtro.CategoriaId)
		argIdx++
	}
	if filtro.Tipo != "" {
		query += fmt.Sprintf(" AND tipo=$%d", argIdx)
		args = append(args, filtro.Tipo)
		argIdx++
	}
	if filtro.Origem != "" {
		query += fmt.Sprintf(" AND origem=$%d", argIdx)
		args = append(args, filtro.Origem)
		argIdx++
	}
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
		args = append(args, filtro.DataInicio)
		argIdx++
	}
	if filtro.DataFim != "" {
		query += fmt.Sprintf(" AND criado_em <= $%d", argIdx)
		args = append(args, filtro.DataFim)
		argIdx++
	}
	if !filtro.PermitirBaseCustomizavel {
		query += " AND tipo != 'base_customizavel'"
	}

	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	var total int
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar templates: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "nome", map[string]string{
		"nome": "nome", "tipo": "tipo", "origem": "origem", "status": "status", "criado_em": "criado_em", "atualizado_em": "atualizado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)

	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar templates: %w", err)
	}
	defer rows.Close()

	var templates []*template.Template
	for rows.Next() {
		t, err := scanner.Template(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, total, nil
}
