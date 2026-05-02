package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/lead"
	"buildnix-backend/internal/infra/database/postgres/internal/scanner"
	"buildnix-backend/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// LeadRepository implementa lead.Repository para PostgreSQL.
type LeadRepository struct {
	connection *Postgres
}

var _ lead.Repository = (*LeadRepository)(nil)

// NewLeadRepository constrói o repositório de leads.
func NewLeadRepository(connection *Postgres) (*LeadRepository, error) {
	return &LeadRepository{connection: connection}, nil
}

func (r *LeadRepository) Salvar(ctx context.Context, l *lead.Lead) error {
	l.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarLead, pgx.NamedArgs{
		"id":                    l.Id,
		"landingpage_id":        l.LandingPageId,
		"nome":                  l.Nome,
		"email":                 l.Email,
		"telefone":              nilableStr(l.Telefone),
		"whatsapp":              nilableStr(l.Whatsapp),
		"mensagem":              nilableStr(l.Mensagem),
		"campos_personalizados": []byte(l.CamposPersonalizados),
		"origem":                l.Origem,
		"ip":                    nilableStr(l.IP),
		"user_agent":            nilableStr(l.UserAgent),
		"status":                string(l.Status),
		"criado_em":             l.CriadoEm,
	})
	if err != nil {
		return fmt.Errorf("salvar lead: %w", err)
	}
	return nil
}

func (r *LeadRepository) BuscarPorID(ctx context.Context, id string) (*lead.Lead, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarLeadPorID, pgx.NamedArgs{"id": id})
	l, err := scanner.Lead(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, lead.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar lead por id: %w", err)
	}
	return l, nil
}

func (r *LeadRepository) AtualizarStatus(ctx context.Context, id string, status lead.Status) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.AtualizarStatusLead, pgx.NamedArgs{
		"status": string(status),
		"id":     id,
	})
	if err != nil {
		return fmt.Errorf("atualizar status lead: %w", err)
	}
	return nil
}

func (r *LeadRepository) Listar(ctx context.Context, filtro lead.FiltroListagem) ([]*lead.Lead, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina

	query := `SELECT l.id, l.landingpage_id, l.nome, l.email, l.telefone, l.whatsapp, l.mensagem,
		l.campos_personalizados, l.origem, l.ip, l.user_agent, l.status, l.criado_em FROM lead l`
	args := []any{}
	argIdx := 1
	if filtro.ClienteId != "" {
		query += " INNER JOIN landingpage lp ON lp.id=l.landingpage_id"
	}
	query += " WHERE 1=1"
	if filtro.ClienteId != "" {
		query += fmt.Sprintf(" AND lp.cliente_id=$%d", argIdx)
		args = append(args, filtro.ClienteId)
		argIdx++
	}

	if filtro.LandingPageId != "" {
		query += fmt.Sprintf(" AND l.landingpage_id=$%d", argIdx)
		args = append(args, filtro.LandingPageId)
		argIdx++
	}
	if filtro.Nome != "" {
		query += fmt.Sprintf(" AND l.nome ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Nome+"%")
		argIdx++
	}
	if filtro.Email != "" {
		query += fmt.Sprintf(" AND l.email ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Email+"%")
		argIdx++
	}
	if filtro.Status != "" {
		query += fmt.Sprintf(" AND l.status=$%d", argIdx)
		args = append(args, filtro.Status)
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND l.criado_em >= $%d", argIdx)
		args = append(args, filtro.DataInicio)
		argIdx++
	}
	if filtro.DataFim != "" {
		query += fmt.Sprintf(" AND l.criado_em <= $%d", argIdx)
		args = append(args, filtro.DataFim)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar leads: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "l.nome", "email": "l.email", "status": "l.status", "criado_em": "l.criado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)

	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar leads: %w", err)
	}
	defer rows.Close()

	var leads []*lead.Lead
	for rows.Next() {
		l, err := scanner.Lead(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear lead: %w", err)
		}
		leads = append(leads, l)
	}
	return leads, total, nil
}
