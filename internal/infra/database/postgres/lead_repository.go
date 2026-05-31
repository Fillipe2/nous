package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/lead"
	"nous/internal/infra/database/postgres/internal/scanner"
	"nous/internal/infra/database/postgres/internal/statement"

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
		"cliente_id":            nilableStr(l.ClienteId),
		"nome":                  l.Nome,
		"email":                 l.Email,
		"telefone":              nilableStr(l.Telefone),
		"whatsapp":              nilableStr(l.Whatsapp),
		"mensagem":              nilableStr(l.Mensagem),
		"campos_personalizados": camposJSON(l.CamposPersonalizados),
		"origem":                l.Origem,
		"ip":                    nilableStr(l.IP),
		"user_agent":            nilableStr(l.UserAgent),
		"referrer":              nilableStr(l.Referrer),
		"utm_source":            nilableStr(l.UTM.Source),
		"utm_medium":            nilableStr(l.UTM.Medium),
		"utm_campaign":          nilableStr(l.UTM.Campaign),
		"utm_content":           nilableStr(l.UTM.Content),
		"utm_term":              nilableStr(l.UTM.Term),
		"consentimento":                         l.Consentimento,
		"consentimento_em":                      l.ConsentimentoEm,
		"politica_versao":                       nilableStr(l.PoliticaVersao),
		"consentimento_marketing_parceiros":      l.ConsentimentoMarketingParceiros,
		"consentimento_versao_marketing":         nilableStr(l.ConsentimentoVersaoMarketing),
		"status":                                string(l.Status),
		"criado_em":             l.CriadoEm,
		"atualizado_em":         l.AtualizadoEm,
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

func (r *LeadRepository) BuscarDuplicado(ctx context.Context, email, landingPageId string, desde time.Time) (*lead.Lead, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarLeadDuplicado, pgx.NamedArgs{
		"email":          email,
		"landingpage_id": landingPageId,
		"desde":          desde,
	})
	l, err := scanner.Lead(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, lead.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar lead duplicado: %w", err)
	}
	return l, nil
}

func (r *LeadRepository) AtualizarStatus(ctx context.Context, id string, atualizacao lead.AtualizacaoStatus, agora time.Time) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.AtualizarStatusLead, pgx.NamedArgs{
		"status":         string(atualizacao.Status),
		"perdido_motivo": nilableStr(atualizacao.Motivo),
		"atualizado_em":  agora,
		"id":             id,
	})
	if err != nil {
		return fmt.Errorf("atualizar status lead: %w", err)
	}
	return nil
}

func (r *LeadRepository) Anonimizar(ctx context.Context, id string, agora time.Time) error {
	stmts := statement.Get()
	tag, err := r.connection.ExecResult(ctx, stmts.AnonimizarLead, pgx.NamedArgs{
		"id":    id,
		"agora": agora,
	})
	if err != nil {
		return fmt.Errorf("anonimizar lead: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return lead.ErrJaAnonimizado
	}
	return nil
}

func (r *LeadRepository) AnonimizarExpirados(ctx context.Context, antesDe time.Time) (int, error) {
	stmts := statement.Get()
	tag, err := r.connection.ExecResult(ctx, stmts.AnonimizarLeadsExpirados, pgx.NamedArgs{
		"antes_de": antesDe,
		"agora":    time.Now().UTC(),
	})
	if err != nil {
		return 0, fmt.Errorf("anonimizar leads expirados: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func (r *LeadRepository) Listar(ctx context.Context, filtro lead.FiltroListagem) ([]*lead.Lead, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina

	query := `SELECT l.id, l.landingpage_id, l.cliente_id, l.nome, l.email, l.telefone, l.whatsapp, l.mensagem,
		l.campos_personalizados, l.origem, l.ip, l.user_agent, l.referrer,
		l.utm_source, l.utm_medium, l.utm_campaign, l.utm_content, l.utm_term,
		l.consentimento, l.consentimento_em, l.politica_versao,
		l.consentimento_marketing_parceiros, l.consentimento_versao_marketing,
		l.status, l.perdido_motivo, l.anonimizado_em, l.criado_em, l.atualizado_em
		FROM lead l`
	args := []any{}
	argIdx := 1
	query += " WHERE 1=1"
	if filtro.ClienteId != "" {
		query += fmt.Sprintf(" AND l.cliente_id=$%d", argIdx)
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
	if filtro.UTMSource != "" {
		query += fmt.Sprintf(" AND l.utm_source=$%d", argIdx)
		args = append(args, filtro.UTMSource)
		argIdx++
	}
	if filtro.UTMCampaign != "" {
		query += fmt.Sprintf(" AND l.utm_campaign=$%d", argIdx)
		args = append(args, filtro.UTMCampaign)
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND l.criado_em >= $%d", argIdx)
		args = append(args, dataInicioFiltro(filtro.DataInicio))
		argIdx++
	}
	if filtro.DataFim != "" {
		operador, valor := dataFimFiltro(filtro.DataFim)
		query += fmt.Sprintf(" AND l.criado_em %s $%d", operador, argIdx)
		args = append(args, valor)
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

func camposJSON(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
