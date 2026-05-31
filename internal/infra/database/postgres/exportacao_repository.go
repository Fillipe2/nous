package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/exportacao"
	"nous/internal/infra/database/postgres/internal/scanner"
	"nous/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// ExportacaoRepository implementa exportacao.Repository para PostgreSQL.
type ExportacaoRepository struct {
	connection *Postgres
}

var _ exportacao.Repository = (*ExportacaoRepository)(nil)

// NewExportacaoRepository constrói o repositório de exportações.
func NewExportacaoRepository(connection *Postgres) (*ExportacaoRepository, error) {
	return &ExportacaoRepository{connection: connection}, nil
}

func (r *ExportacaoRepository) Salvar(ctx context.Context, e *exportacao.Exportacao) error {
	e.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarExportacao, pgx.NamedArgs{
		"id":             e.Id,
		"landingpage_id": e.LandingPageId,
		"usuario_id":     e.UsuarioId,
		"nome_arquivo":   e.NomeArquivo,
		"content_type":   e.ContentType,
		"tamanho_bytes":  e.TamanhoBytes,
		"arquivo":        e.Arquivo,
		"hash":           e.Hash,
		"criado_em":      e.CriadoEm,
	})
	if err != nil {
		return fmt.Errorf("salvar exportacao: %w", err)
	}
	return nil
}

func (r *ExportacaoRepository) BuscarPorID(ctx context.Context, id string) (*exportacao.Exportacao, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarExportacaoPorID, pgx.NamedArgs{"id": id})
	e, err := scanner.Exportacao(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, exportacao.ErrNaoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("buscar exportacao por id: %w", err)
	}
	return e, nil
}

func (r *ExportacaoRepository) BuscarMaisRecentePorLandingPage(ctx context.Context, landingPageId string) (*exportacao.Exportacao, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarMaisRecenteExportacaoPorLandingPage, pgx.NamedArgs{"landingpage_id": landingPageId})
	e, err := scanner.Exportacao(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, exportacao.ErrNaoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("buscar exportacao mais recente: %w", err)
	}
	return e, nil
}

func (r *ExportacaoRepository) Listar(ctx context.Context, filtro exportacao.FiltroListagem) ([]*exportacao.Exportacao, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina
	query := `SELECT e.id, e.landingpage_id, e.usuario_id, e.nome_arquivo, e.content_type, e.tamanho_bytes, ''::bytea AS arquivo, e.hash, e.criado_em FROM exportacao e`
	args := []any{}
	argIdx := 1
	if filtro.ClienteId != "" {
		query += " INNER JOIN landingpage lp ON lp.id=e.landingpage_id"
	}
	query += " WHERE 1=1"
	if filtro.ClienteId != "" {
		query += fmt.Sprintf(" AND lp.cliente_id=$%d", argIdx)
		args = append(args, filtro.ClienteId)
		argIdx++
	}
	if filtro.LandingPageId != "" {
		query += fmt.Sprintf(" AND e.landingpage_id=$%d", argIdx)
		args = append(args, filtro.LandingPageId)
		argIdx++
	}
	if filtro.UsuarioId != "" {
		query += fmt.Sprintf(" AND e.usuario_id=$%d", argIdx)
		args = append(args, filtro.UsuarioId)
		argIdx++
	}
	if filtro.NomeArquivo != "" {
		query += fmt.Sprintf(" AND e.nome_arquivo ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.NomeArquivo+"%")
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND e.criado_em >= $%d", argIdx)
		args = append(args, dataInicioFiltro(filtro.DataInicio))
		argIdx++
	}
	if filtro.DataFim != "" {
		operador, valor := dataFimFiltro(filtro.DataFim)
		query += fmt.Sprintf(" AND e.criado_em %s $%d", operador, argIdx)
		args = append(args, valor)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	var total int
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar exportacoes: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome_arquivo": "e.nome_arquivo", "tamanho_bytes": "e.tamanho_bytes", "criado_em": "e.criado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)
	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar exportacoes: %w", err)
	}
	defer rows.Close()

	var exportacoes []*exportacao.Exportacao
	for rows.Next() {
		e, err := scanner.Exportacao(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear exportacao: %w", err)
		}
		exportacoes = append(exportacoes, e)
	}
	return exportacoes, total, nil
}
