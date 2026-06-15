package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/landingpage"
	"nous/internal/infra/database/postgres/internal/scanner"
	"nous/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// LandingPageRepository implementa landingpage.Repository para PostgreSQL.
type LandingPageRepository struct {
	connection *Postgres
}

var _ landingpage.Repository = (*LandingPageRepository)(nil)

// NewLandingPageRepository constrói o repositório de landing pages.
func NewLandingPageRepository(connection *Postgres) (*LandingPageRepository, error) {
	return &LandingPageRepository{connection: connection}, nil
}

func (r *LandingPageRepository) Salvar(ctx context.Context, lp *landingpage.LandingPage) error {
	lp.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarLandingPage, pgx.NamedArgs{
		"id":            lp.Id,
		"negocio_id":    lp.NegocioId,
		"unidade_id":    nilableStr(lp.UnidadeId),
		"categoria_id":  nilableStr(lp.CategoriaId),
		"template_id":   nilableStr(lp.TemplateId),
		"tipo":          string(lp.Tipo),
		"nome":          lp.Nome,
		"slug":          nilableStr(lp.Slug),
		"conteudo":      nilableStr(lp.Conteudo),
		"status":        string(lp.Status),
		"is_draft":      lp.IsDraft,
		"autosave_data": nilableJSON(lp.AutosaveData),
		"criado_em":     lp.CriadoEm,
		"atualizado_em": lp.AtualizadoEm,
		"version":       lp.Version,
	})
	if err != nil {
		if isUniqueViolation(err, "idx_landingpage_slug") {
			return landingpage.ErrSlugJaExiste
		}
		return fmt.Errorf("salvar landingpage: %w", err)
	}
	return nil
}

func (r *LandingPageRepository) BuscarPorID(ctx context.Context, id string) (*landingpage.LandingPage, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarLandingPagePorID, pgx.NamedArgs{"id": id})
	lp, err := scanner.LandingPage(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, landingpage.ErrNaoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("buscar landingpage por id: %w", err)
	}
	return lp, nil
}

func (r *LandingPageRepository) Atualizar(ctx context.Context, lp *landingpage.LandingPage) error {
	stmts := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, stmts.AtualizarLandingPage, pgx.NamedArgs{
		"unidade_id":       nilableStr(lp.UnidadeId),
		"categoria_id":     nilableStr(lp.CategoriaId),
		"template_id":      nilableStr(lp.TemplateId),
		"tipo":             string(lp.Tipo),
		"nome":             lp.Nome,
		"slug":             nilableStr(lp.Slug),
		"conteudo":         nilableStr(lp.Conteudo),
		"status":           string(lp.Status),
		"is_draft":         lp.IsDraft,
		"atualizado_em":    lp.AtualizadoEm,
		"version":          lp.Version,
		"version_anterior": lp.Version - 1,
		"id":               lp.Id,
	})
	if err != nil {
		if isUniqueViolation(err, "idx_landingpage_slug") {
			return landingpage.ErrSlugJaExiste
		}
		return fmt.Errorf("atualizar landingpage: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return landingpage.ErrVersaoConflito
	}
	return nil
}

func (r *LandingPageRepository) SoftDelete(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SoftDeleteLandingPage, pgx.NamedArgs{
		"atualizado_em": time.Now(),
		"id":            id,
	})
	if err != nil {
		return fmt.Errorf("soft delete landingpage: %w", err)
	}
	return nil
}

func (r *LandingPageRepository) AtualizarAutosave(ctx context.Context, id string, dados []byte, version int) (int, error) {
	stmts := statement.Get()
	var novaVersao int
	err := r.connection.QueryRow(ctx, stmts.AtualizarAutosaveLandingPage, pgx.NamedArgs{
		"autosave_data":    string(dados),
		"last_autosave_em": time.Now(),
		"version":          version,
		"id":               id,
	}).Scan(&novaVersao)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, landingpage.ErrVersaoConflito
	}
	if err != nil {
		return 0, fmt.Errorf("autosave landingpage: %w", err)
	}
	return novaVersao, nil
}

func (r *LandingPageRepository) AtualizarLock(ctx context.Context, id string, usuarioId *string) error {
	stmts := statement.Get()
	if usuarioId != nil {
		agora := time.Now()
		commandTag, err := r.connection.ExecResult(ctx, stmts.BloquearLandingPage, pgx.NamedArgs{
			"locked_by":   *usuarioId,
			"locked_at":   agora,
			"expirado_em": agora.Add(-landingpage.DuracaoLock),
			"id":          id,
		})
		if err != nil {
			return fmt.Errorf("bloquear landingpage: %w", err)
		}
		if commandTag.RowsAffected() == 0 {
			return landingpage.ErrBloqueada
		}
	} else {
		err := r.connection.Exec(ctx, stmts.DesbloquearLandingPage, pgx.NamedArgs{"id": id})
		if err != nil {
			return fmt.Errorf("desbloquear landingpage: %w", err)
		}
	}
	return nil
}

func (r *LandingPageRepository) VerificarAcessoUsuario(ctx context.Context, landingPageId, usuarioId string) error {
	stmts := statement.Get()
	var acesso bool
	if err := r.connection.QueryRow(ctx, stmts.VerificarAcessoLandingPage, pgx.NamedArgs{
		"id":         landingPageId,
		"usuario_id": usuarioId,
	}).Scan(&acesso); err != nil {
		return fmt.Errorf("verificar acesso usuario: %w", err)
	}
	if !acesso {
		return landingpage.ErrNaoAutorizado
	}
	return nil
}

func (r *LandingPageRepository) ContarPorNegocio(ctx context.Context, negocioId string) (int, error) {
	stmts := statement.Get()
	var total int
	if err := r.connection.QueryRow(ctx, stmts.ContarLandingPagesPorNegocio, pgx.NamedArgs{"negocio_id": negocioId}).Scan(&total); err != nil {
		return 0, fmt.Errorf("contar landingpages por negocio: %w", err)
	}
	return total, nil
}

func (r *LandingPageRepository) Listar(ctx context.Context, filtro landingpage.FiltroListagem) ([]*landingpage.LandingPage, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina

	query := `SELECT lp.id, lp.negocio_id, lp.unidade_id, lp.categoria_id, lp.template_id, lp.tipo,
		lp.nome, lp.conteudo, lp.status, lp.is_draft, lp.autosave_data,
		lp.last_autosave_em, lp.locked_by, lp.locked_at, lp.criado_em, lp.atualizado_em, lp.version,
		lp.slug, lp.url_publica, lp.deploy_id, lp.publicada_em
		FROM landingpage lp`
	args := []any{}
	argIdx := 1

	if filtro.UsuarioId != "" {
		query += fmt.Sprintf(" INNER JOIN usuario_negocio uc ON uc.negocio_id=lp.negocio_id AND uc.usuario_id=$%d", argIdx)
		args = append(args, filtro.UsuarioId)
		argIdx++
	}

	query += " WHERE lp.status != 'excluida'"

	if filtro.NegocioId != "" {
		query += fmt.Sprintf(" AND lp.negocio_id=$%d", argIdx)
		args = append(args, filtro.NegocioId)
		argIdx++
	}
	if filtro.UnidadeId != "" {
		query += fmt.Sprintf(" AND lp.unidade_id=$%d", argIdx)
		args = append(args, filtro.UnidadeId)
		argIdx++
	}
	if filtro.CategoriaId != "" {
		query += fmt.Sprintf(" AND lp.categoria_id=$%d", argIdx)
		args = append(args, filtro.CategoriaId)
		argIdx++
	}
	if filtro.TemplateId != "" {
		query += fmt.Sprintf(" AND lp.template_id=$%d", argIdx)
		args = append(args, filtro.TemplateId)
		argIdx++
	}
	if filtro.Nome != "" {
		query += fmt.Sprintf(" AND lp.nome ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Nome+"%")
		argIdx++
	}
	if filtro.Tipo != "" {
		query += fmt.Sprintf(" AND lp.tipo=$%d", argIdx)
		args = append(args, filtro.Tipo)
		argIdx++
	}
	if filtro.Status != "" {
		query += fmt.Sprintf(" AND lp.status=$%d", argIdx)
		args = append(args, filtro.Status)
		argIdx++
	}
	if filtro.IsDraft == "true" || filtro.IsDraft == "false" {
		query += fmt.Sprintf(" AND lp.is_draft=$%d", argIdx)
		args = append(args, filtro.IsDraft == "true")
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND lp.criado_em >= $%d", argIdx)
		args = append(args, dataInicioFiltro(filtro.DataInicio))
		argIdx++
	}
	if filtro.DataFim != "" {
		operador, valor := dataFimFiltro(filtro.DataFim)
		query += fmt.Sprintf(" AND lp.criado_em %s $%d", operador, argIdx)
		args = append(args, valor)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar landingpages: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "lp.nome", "tipo": "lp.tipo", "status": "lp.status", "criado_em": "lp.criado_em", "atualizado_em": "lp.atualizado_em", "last_autosave_em": "lp.last_autosave_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)

	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar landingpages: %w", err)
	}
	defer rows.Close()

	var landingPages []*landingpage.LandingPage
	for rows.Next() {
		lp, err := scanner.LandingPage(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear landingpage: %w", err)
		}
		landingPages = append(landingPages, lp)
	}
	return landingPages, total, nil
}

func (r *LandingPageRepository) AtualizarPublicacao(ctx context.Context, id, urlPublica, deployId string, publicadaEm time.Time) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.AtualizarPublicacaoLandingPage, pgx.NamedArgs{
		"id":            id,
		"url_publica":   urlPublica,
		"deploy_id":     deployId,
		"publicada_em":  publicadaEm,
		"atualizado_em": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("atualizar publicacao landingpage: %w", err)
	}
	return nil
}

func (r *LandingPageRepository) AtualizarDespublicacao(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.AtualizarDespublicacaoLandingPage, pgx.NamedArgs{
		"id":            id,
		"atualizado_em": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("atualizar despublicacao landingpage: %w", err)
	}
	return nil
}

func nilableJSON(dados []byte) *string {
	if len(dados) == 0 {
		return nil
	}
	valor := string(dados)
	return &valor
}
