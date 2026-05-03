package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/auditoria"
	"buildnix-backend/internal/infra/database/postgres/internal/scanner"
	"buildnix-backend/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// AuditoriaRepository implementa auditoria.Repository para PostgreSQL.
type AuditoriaRepository struct {
	connection *Postgres
}

var _ auditoria.Repository = (*AuditoriaRepository)(nil)

// NewAuditoriaRepository constrói o repositório de auditoria.
func NewAuditoriaRepository(connection *Postgres) (*AuditoriaRepository, error) {
	return &AuditoriaRepository{connection: connection}, nil
}

func (r *AuditoriaRepository) Salvar(ctx context.Context, a *auditoria.Auditoria) error {
	a.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarAuditoria, pgx.NamedArgs{
		"id": a.Id,
		"usuario_id": nilableStr(func() string {
			if a.UsuarioId != nil {
				return *a.UsuarioId
			}
			return ""
		}()),
		"acao":              a.Acao,
		"entidade":          nilableStr(a.Entidade),
		"entidade_id":       a.EntidadeId,
		"dados_anteriores":  []byte(a.DadosAnteriores),
		"dados_novos":       []byte(a.DadosNovos),
		"ip":                nilableStr(a.IP),
		"user_agent":        nilableStr(a.UserAgent),
		"status_code":       a.StatusCode,
		"tempo_resposta_ms": a.TempoRespostaMs,
		"request_id":        nilableStr(a.RequestId),
		"criado_em":         a.CriadoEm,
	})
	if err != nil {
		return fmt.Errorf("salvar auditoria: %w", err)
	}
	return nil
}

func (r *AuditoriaRepository) BuscarPorID(ctx context.Context, id string) (*auditoria.Auditoria, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarAuditoriaPorID, pgx.NamedArgs{"id": id})
	a, err := scanner.Auditoria(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("auditoria nao encontrada")
	}
	if err != nil {
		return nil, fmt.Errorf("buscar auditoria por id: %w", err)
	}
	return a, nil
}

func (r *AuditoriaRepository) Listar(ctx context.Context, filtro auditoria.FiltroListagem) ([]*auditoria.Auditoria, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina

	query := `SELECT id, usuario_id, acao, entidade, entidade_id, dados_anteriores, dados_novos, ip, user_agent, status_code, tempo_resposta_ms, request_id, criado_em
		FROM auditoria WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filtro.UsuarioId != "" {
		query += fmt.Sprintf(" AND usuario_id=$%d", argIdx)
		args = append(args, filtro.UsuarioId)
		argIdx++
	}
	if filtro.Acao != "" {
		query += fmt.Sprintf(" AND acao ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Acao+"%")
		argIdx++
	}
	if filtro.Entidade != "" {
		query += fmt.Sprintf(" AND entidade=$%d", argIdx)
		args = append(args, filtro.Entidade)
		argIdx++
	}
	if filtro.EntidadeId != "" {
		query += fmt.Sprintf(" AND entidade_id=$%d", argIdx)
		args = append(args, filtro.EntidadeId)
		argIdx++
	}
	if filtro.StatusCode != "" {
		query += fmt.Sprintf(" AND status_code::text=$%d", argIdx)
		args = append(args, filtro.StatusCode)
		argIdx++
	}
	if filtro.RequestId != "" {
		query += fmt.Sprintf(" AND request_id=$%d", argIdx)
		args = append(args, filtro.RequestId)
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
		return nil, 0, fmt.Errorf("contar auditorias: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"acao": "acao", "entidade": "entidade", "status_code": "status_code", "tempo_resposta_ms": "tempo_resposta_ms", "criado_em": "criado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)

	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar auditorias: %w", err)
	}
	defer rows.Close()

	var auditorias []*auditoria.Auditoria
	for rows.Next() {
		a, err := scanner.Auditoria(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear auditoria: %w", err)
		}
		auditorias = append(auditorias, a)
	}
	return auditorias, total, nil
}
