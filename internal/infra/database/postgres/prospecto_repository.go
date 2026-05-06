package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/prospecto"
	"buildnix-backend/internal/infra/database/postgres/internal/scanner"
	"buildnix-backend/internal/infra/database/postgres/internal/statement"
)

// ProspectoRepository implementa prospecto.Repository para PostgreSQL.
type ProspectoRepository struct {
	connection *Postgres
}

var _ prospecto.Repository = (*ProspectoRepository)(nil)

// NewProspectoRepository constrói o repositório de prospectos.
func NewProspectoRepository(connection *Postgres) (*ProspectoRepository, error) {
	return &ProspectoRepository{connection: connection}, nil
}

func (r *ProspectoRepository) Salvar(ctx context.Context, p *prospecto.Prospecto) error {
	p.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarProspecto, pgx.NamedArgs{
		"id":                         p.Id,
		"nome":                       p.Nome,
		"email":                      p.Email,
		"telefone":                   nilableStr(p.Telefone),
		"categoria_pretendida_id":    nilableStr(p.CategoriaPretendidaId),
		"categoria_pretendida_livre": nilableStr(p.CategoriaPretendidaLivre),
		"cidade":                     nilableStr(p.Cidade),
		"estado":                     nilableStr(p.Estado),
		"porte":                      nilableStr(string(p.Porte)),
		"interesses":                 camposJSON(p.Interesses),
		"mensagem":                   nilableStr(p.Mensagem),
		"origem":                     nilableStr(p.Origem),
		"utm_source":                 nilableStr(p.UTMSource),
		"utm_medium":                 nilableStr(p.UTMMedium),
		"utm_campaign":               nilableStr(p.UTMCampaign),
		"utm_content":                nilableStr(p.UTMContent),
		"utm_term":                   nilableStr(p.UTMTerm),
		"status_funil":               string(p.StatusFunil),
		"convertido_em_usuario_id":   nilableStr(p.ConvertidoEmUsuarioId),
		"ip":                         nilableStr(p.IP),
		"user_agent":                 nilableStr(p.UserAgent),
		"criado_em":                  p.CriadoEm,
		"atualizado_em":              p.AtualizadoEm,
	})
	if err != nil {
		return fmt.Errorf("salvar prospecto: %w", err)
	}
	return nil
}

func (r *ProspectoRepository) BuscarPorID(ctx context.Context, id string) (*prospecto.Prospecto, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarProspectoPorID, pgx.NamedArgs{"id": id})
	p, err := scanner.Prospecto(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, prospecto.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar prospecto por id: %w", err)
	}
	return p, nil
}

func (r *ProspectoRepository) BuscarPorEmail(ctx context.Context, email string) (*prospecto.Prospecto, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarProspectoPorEmail, pgx.NamedArgs{"email": email})
	p, err := scanner.Prospecto(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, prospecto.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar prospecto por email: %w", err)
	}
	return p, nil
}

func (r *ProspectoRepository) AtualizarFunil(ctx context.Context, id string, atualizacao prospecto.AtualizacaoFunil, agora time.Time) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.AtualizarFunilProspecto, pgx.NamedArgs{
		"status_funil":             string(atualizacao.StatusFunil),
		"convertido_em_usuario_id": nilableStr(atualizacao.ConvertidoEmUsuarioId),
		"atualizado_em":            agora,
		"id":                       id,
	})
	if err != nil {
		return fmt.Errorf("atualizar funil prospecto: %w", err)
	}
	return nil
}

func (r *ProspectoRepository) Listar(ctx context.Context, filtro prospecto.FiltroListagem) ([]*prospecto.Prospecto, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina

	query := `SELECT id, nome, email, telefone,
		categoria_pretendida_id, categoria_pretendida_livre,
		cidade, estado, porte, interesses, mensagem, origem,
		utm_source, utm_medium, utm_campaign, utm_content, utm_term,
		status_funil, convertido_em_usuario_id,
		ip, user_agent, criado_em, atualizado_em
		FROM prospecto`
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
		return nil, 0, fmt.Errorf("contar prospectos: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "nome", "email": "email", "status_funil": "status_funil", "criado_em": "criado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)

	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar prospectos: %w", err)
	}
	defer rows.Close()

	var prospectos []*prospecto.Prospecto
	for rows.Next() {
		p, err := scanner.Prospecto(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear prospecto: %w", err)
		}
		prospectos = append(prospectos, p)
	}
	return prospectos, total, nil
}
