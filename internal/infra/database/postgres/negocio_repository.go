package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/negocio"
	"nous/internal/infra/database/postgres/internal/scanner"
	"nous/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// NegocioRepository implementa negocio.Repository para PostgreSQL.
type NegocioRepository struct {
	connection *Postgres
}

var _ negocio.Repository = (*NegocioRepository)(nil)

// NewNegocioRepository constrói o repositório de negocios.
func NewNegocioRepository(connection *Postgres) (*NegocioRepository, error) {
	return &NegocioRepository{connection: connection}, nil
}

// documentoArg grava NULL quando o documento está vazio, evitando strings vazias
// na coluna opcional.
func documentoArg(doc string) any {
	if doc == "" {
		return nil
	}
	return doc
}

func (r *NegocioRepository) Salvar(ctx context.Context, c *negocio.Negocio) error {
	c.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarNegocio, pgx.NamedArgs{
		"id":            c.Id,
		"nome":          c.Nome,
		"email":         c.Email,
		"telefone":      c.Telefone,
		"whatsapp":      c.Whatsapp,
		"documento":     documentoArg(c.Documento),
		"status":        string(c.Status),
		"criado_em":     c.CriadoEm,
		"atualizado_em": c.AtualizadoEm,
		"version":       c.Version,
	})
	if err != nil {
		return fmt.Errorf("salvar negocio: %w", err)
	}
	return nil
}

// SalvarComVinculo cria o negocio e o vínculo usuário-negocio numa única
// transação, evitando negocios órfãos caso o vínculo falhe.
func (r *NegocioRepository) SalvarComVinculo(ctx context.Context, c *negocio.Negocio, uc *negocio.UsuarioNegocio) error {
	c.Id = uuid.New().String()
	uc.NegocioId = c.Id
	stmts := statement.Get()

	tx, err := r.connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciar transacao salvar negocio: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, stmts.SalvarNegocio, pgx.NamedArgs{
		"id":            c.Id,
		"nome":          c.Nome,
		"email":         c.Email,
		"telefone":      c.Telefone,
		"whatsapp":      c.Whatsapp,
		"documento":     documentoArg(c.Documento),
		"status":        string(c.Status),
		"criado_em":     c.CriadoEm,
		"atualizado_em": c.AtualizadoEm,
		"version":       c.Version,
	}); err != nil {
		return fmt.Errorf("salvar negocio: %w", err)
	}

	if _, err := tx.Exec(ctx, stmts.VincularUsuarioNegocio, pgx.NamedArgs{
		"usuario_id": uc.UsuarioId,
		"negocio_id": uc.NegocioId,
		"criado_em":  uc.CriadoEm,
	}); err != nil {
		return fmt.Errorf("vincular usuario negocio: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit salvar negocio: %w", err)
	}
	return nil
}

func (r *NegocioRepository) BuscarPorID(ctx context.Context, id string) (*negocio.Negocio, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarNegocioPorID, pgx.NamedArgs{"id": id})
	c, err := scanner.Negocio(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, negocio.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar negocio por id: %w", err)
	}
	return c, nil
}

func (r *NegocioRepository) Atualizar(ctx context.Context, c *negocio.Negocio) error {
	stmts := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, stmts.AtualizarNegocio, pgx.NamedArgs{
		"nome":             c.Nome,
		"email":            c.Email,
		"telefone":         c.Telefone,
		"whatsapp":         c.Whatsapp,
		"documento":        documentoArg(c.Documento),
		"status":           string(c.Status),
		"atualizado_em":    c.AtualizadoEm,
		"version":          c.Version,
		"version_anterior": c.Version - 1,
		"id":               c.Id,
	})
	if err != nil {
		return fmt.Errorf("atualizar negocio: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return negocio.ErrVersaoConflito
	}
	return nil
}

func (r *NegocioRepository) SoftDelete(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SoftDeleteNegocio, pgx.NamedArgs{
		"atualizado_em": time.Now(),
		"id":            id,
	})
	if err != nil {
		return fmt.Errorf("soft delete negocio: %w", err)
	}
	return nil
}

func (r *NegocioRepository) Listar(ctx context.Context, filtro negocio.FiltroListagem) ([]*negocio.Negocio, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina
	query := `SELECT c.id, c.nome, c.email, c.telefone, c.whatsapp, c.documento, c.status, c.criado_em, c.atualizado_em, c.version FROM negocio c`
	args := []any{}
	argIdx := 1
	if filtro.UsuarioId != "" {
		query += " INNER JOIN usuario_negocio uc ON uc.negocio_id=c.id"
	}
	query += " WHERE c.status != 'excluido'"
	if filtro.UsuarioId != "" {
		query += fmt.Sprintf(" AND uc.usuario_id=$%d", argIdx)
		args = append(args, filtro.UsuarioId)
		argIdx++
	}

	if filtro.Nome != "" {
		query += fmt.Sprintf(" AND c.nome ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Nome+"%")
		argIdx++
	}
	if filtro.Email != "" {
		query += fmt.Sprintf(" AND c.email ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Email+"%")
		argIdx++
	}
	if filtro.Status != "" {
		query += fmt.Sprintf(" AND c.status=$%d", argIdx)
		args = append(args, filtro.Status)
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND c.criado_em >= $%d", argIdx)
		args = append(args, dataInicioFiltro(filtro.DataInicio))
		argIdx++
	}
	if filtro.DataFim != "" {
		operador, valor := dataFimFiltro(filtro.DataFim)
		query += fmt.Sprintf(" AND c.criado_em %s $%d", operador, argIdx)
		args = append(args, valor)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar negocios: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "c.nome", "email": "c.email", "status": "c.status", "criado_em": "c.criado_em", "atualizado_em": "c.atualizado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)
	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar negocios: %w", err)
	}
	defer rows.Close()

	var negocios []*negocio.Negocio
	for rows.Next() {
		c, err := scanner.Negocio(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear negocio: %w", err)
		}
		negocios = append(negocios, c)
	}
	return negocios, total, nil
}

func (r *NegocioRepository) VincularUsuario(ctx context.Context, uc *negocio.UsuarioNegocio) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.VincularUsuarioNegocio, pgx.NamedArgs{
		"usuario_id": uc.UsuarioId,
		"negocio_id": uc.NegocioId,
		"criado_em":  uc.CriadoEm,
	})
	if err != nil {
		return fmt.Errorf("vincular usuario negocio: %w", err)
	}
	return nil
}

func (r *NegocioRepository) ListarPorUsuario(ctx context.Context, usuarioId string) ([]*negocio.Negocio, error) {
	stmts := statement.Get()
	rows, err := r.connection.Query(ctx, stmts.ListarNegociosPorUsuario, pgx.NamedArgs{"usuario_id": usuarioId})
	if err != nil {
		return nil, fmt.Errorf("listar negocios por usuario: %w", err)
	}
	defer rows.Close()

	var negocios []*negocio.Negocio
	for rows.Next() {
		c, err := scanner.Negocio(rows)
		if err != nil {
			return nil, fmt.Errorf("escanear negocio: %w", err)
		}
		negocios = append(negocios, c)
	}
	return negocios, nil
}

func (r *NegocioRepository) VerificarAcessoUsuario(ctx context.Context, negocioId, usuarioId string) error {
	stmts := statement.Get()
	var acesso bool
	if err := r.connection.QueryRow(ctx, stmts.VerificarAcessoNegocio, pgx.NamedArgs{
		"negocio_id": negocioId,
		"usuario_id": usuarioId,
	}).Scan(&acesso); err != nil {
		return fmt.Errorf("verificar acesso negocio: %w", err)
	}
	if !acesso {
		return negocio.ErrNaoAutorizado
	}
	return nil
}
