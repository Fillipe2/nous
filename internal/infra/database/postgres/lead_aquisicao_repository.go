package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "buildnix-backend/internal/domain/lead_aquisicao"
	"buildnix-backend/internal/infra/database/postgres/internal/scanner"
	"buildnix-backend/internal/infra/database/postgres/internal/statement"
)

// LeadAquisicaoRepository implementa lead_aquisicao.Repository para PostgreSQL.
type LeadAquisicaoRepository struct {
	connection *Postgres
}

var _ domain.Repository = (*LeadAquisicaoRepository)(nil)

// NewLeadAquisicaoRepository constrói o repositório de leads de aquisição.
func NewLeadAquisicaoRepository(connection *Postgres) (*LeadAquisicaoRepository, error) {
	return &LeadAquisicaoRepository{connection: connection}, nil
}

func (r *LeadAquisicaoRepository) Salvar(ctx context.Context, la *domain.LeadAquisicao) error {
	la.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarLeadAquisicao, pgx.NamedArgs{
		"id":                         la.Id,
		"nome":                       la.Nome,
		"email":                      la.Email,
		"telefone":                   nilableStr(la.Telefone),
		"categoria_pretendida_id":    nilableStr(la.CategoriaPretendidaId),
		"categoria_pretendida_livre": nilableStr(la.CategoriaPretendidaLivre),
		"cidade":                     nilableStr(la.Cidade),
		"estado":                     nilableStr(la.Estado),
		"porte":                      nilableStr(string(la.Porte)),
		"interesses":                 camposJSON(la.Interesses),
		"mensagem":                   nilableStr(la.Mensagem),
		"origem":                     nilableStr(la.Origem),
		"utm_source":                 nilableStr(la.UTMSource),
		"utm_medium":                 nilableStr(la.UTMMedium),
		"utm_campaign":               nilableStr(la.UTMCampaign),
		"utm_content":                nilableStr(la.UTMContent),
		"utm_term":                   nilableStr(la.UTMTerm),
		"status_funil":               string(la.StatusFunil),
		"convertido_em_usuario_id":   nilableStr(la.ConvertidoEmUsuarioId),
		"ip":                         nilableStr(la.IP),
		"user_agent":                 nilableStr(la.UserAgent),
		"criado_em":                  la.CriadoEm,
		"atualizado_em":              la.AtualizadoEm,
	})
	if err != nil {
		return fmt.Errorf("salvar lead_aquisicao: %w", err)
	}
	return nil
}

func (r *LeadAquisicaoRepository) BuscarPorID(ctx context.Context, id string) (*domain.LeadAquisicao, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarLeadAquisicaoPorID, pgx.NamedArgs{"id": id})
	la, err := scanner.LeadAquisicao(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar lead_aquisicao por id: %w", err)
	}
	return la, nil
}

func (r *LeadAquisicaoRepository) BuscarPorEmail(ctx context.Context, email string) (*domain.LeadAquisicao, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarLeadAquisicaoPorEmail, pgx.NamedArgs{"email": email})
	la, err := scanner.LeadAquisicao(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar lead_aquisicao por email: %w", err)
	}
	return la, nil
}

func (r *LeadAquisicaoRepository) AtualizarFunil(ctx context.Context, id string, atualizacao domain.AtualizacaoFunil, agora time.Time) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.AtualizarFunilLeadAquisicao, pgx.NamedArgs{
		"status_funil":             string(atualizacao.StatusFunil),
		"convertido_em_usuario_id": nilableStr(atualizacao.ConvertidoEmUsuarioId),
		"atualizado_em":            agora,
		"id":                       id,
	})
	if err != nil {
		return fmt.Errorf("atualizar funil lead_aquisicao: %w", err)
	}
	return nil
}

func (r *LeadAquisicaoRepository) Listar(ctx context.Context, filtro domain.FiltroListagem) ([]*domain.LeadAquisicao, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina

	query := `SELECT id, nome, email, telefone,
		categoria_pretendida_id, categoria_pretendida_livre,
		cidade, estado, porte, interesses, mensagem, origem,
		utm_source, utm_medium, utm_campaign, utm_content, utm_term,
		status_funil, convertido_em_usuario_id,
		ip, user_agent, criado_em, atualizado_em
		FROM lead_aquisicao`
	args := []any{}
	argIdx := 1
	query += " WHERE 1=1"

	if filtro.Nome != "" {
		query += fmt.Sprintf(" AND nome ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Nome+"%")
		argIdx++
	}
	if filtro.Email != "" {
		query += fmt.Sprintf(" AND email ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Email+"%")
		argIdx++
	}
	if filtro.StatusFunil != "" {
		query += fmt.Sprintf(" AND status_funil=$%d", argIdx)
		args = append(args, filtro.StatusFunil)
		argIdx++
	}
	if filtro.Porte != "" {
		query += fmt.Sprintf(" AND porte=$%d", argIdx)
		args = append(args, filtro.Porte)
		argIdx++
	}
	if filtro.Cidade != "" {
		query += fmt.Sprintf(" AND cidade ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Cidade+"%")
		argIdx++
	}
	if filtro.Estado != "" {
		query += fmt.Sprintf(" AND estado=$%d", argIdx)
		args = append(args, filtro.Estado)
		argIdx++
	}
	if filtro.UTMCampaign != "" {
		query += fmt.Sprintf(" AND utm_campaign=$%d", argIdx)
		args = append(args, filtro.UTMCampaign)
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
		return nil, 0, fmt.Errorf("contar leads_aquisicao: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "nome", "email": "email", "status_funil": "status_funil", "criado_em": "criado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)

	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar leads_aquisicao: %w", err)
	}
	defer rows.Close()

	var leads []*domain.LeadAquisicao
	for rows.Next() {
		la, err := scanner.LeadAquisicao(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear lead_aquisicao: %w", err)
		}
		leads = append(leads, la)
	}
	return leads, total, nil
}
