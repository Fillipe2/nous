package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/usuario"
	"buildnix-backend/internal/infra/database/postgres/internal/scanner"
	"buildnix-backend/internal/infra/database/postgres/internal/statement"

	"github.com/google/uuid"
)

// UsuarioRepository implementa usuario.Repository para PostgreSQL.
type UsuarioRepository struct {
	connection *Postgres
}

var _ usuario.Repository = (*UsuarioRepository)(nil)

// NewUsuarioRepository constrói o repositório de usuários.
func NewUsuarioRepository(connection *Postgres) (*UsuarioRepository, error) {
	return &UsuarioRepository{connection: connection}, nil
}

func (r *UsuarioRepository) Salvar(ctx context.Context, u *usuario.Usuario) error {
	u.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarUsuario, pgx.NamedArgs{
		"id":            u.Id,
		"nome":          u.Nome,
		"email":         u.Email,
		"senha_hash":    nilableStr(u.SenhaHash),
		"role":          string(u.Role),
		"status":        string(u.Status),
		"criado_em":     u.CriadoEm,
		"atualizado_em": u.AtualizadoEm,
		"version":       u.Version,
	})
	if err != nil {
		return fmt.Errorf("salvar usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) BuscarPorID(ctx context.Context, id string) (*usuario.Usuario, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarUsuarioPorID, pgx.NamedArgs{"id": id})
	u, err := scanner.Usuario(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usuario.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar usuario por id: %w", err)
	}
	return u, nil
}

func (r *UsuarioRepository) BuscarPorEmail(ctx context.Context, email string) (*usuario.Usuario, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarUsuarioPorEmail, pgx.NamedArgs{"email": email})
	u, err := scanner.Usuario(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usuario.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar usuario por email: %w", err)
	}
	return u, nil
}

func (r *UsuarioRepository) Atualizar(ctx context.Context, u *usuario.Usuario) error {
	stmts := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, stmts.AtualizarUsuario, pgx.NamedArgs{
		"nome":             u.Nome,
		"email":            u.Email,
		"senha_hash":       nilableStr(u.SenhaHash),
		"role":             string(u.Role),
		"status":           string(u.Status),
		"atualizado_em":    u.AtualizadoEm,
		"version":          u.Version,
		"version_anterior": u.Version - 1,
		"id":               u.Id,
	})
	if err != nil {
		return fmt.Errorf("atualizar usuario: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return usuario.ErrVersaoConflito
	}
	return nil
}

func (r *UsuarioRepository) SoftDelete(ctx context.Context, id string) error {
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SoftDeleteUsuario, pgx.NamedArgs{
		"atualizado_em": time.Now(),
		"id":            id,
	})
	if err != nil {
		return fmt.Errorf("soft delete usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) Listar(ctx context.Context, filtro usuario.FiltroListagem) ([]*usuario.Usuario, int, error) {
	offset := (filtro.Pagina - 1) * filtro.PorPagina
	query := `SELECT u.id, u.nome, u.email, u.senha_hash, u.role, u.status, u.criado_em, u.atualizado_em, u.version FROM usuario u`
	args := []any{}
	argIdx := 1
	if filtro.ClienteId != "" {
		query += " INNER JOIN usuario_cliente uc ON uc.usuario_id=u.id"
	}
	query += " WHERE 1=1"
	if filtro.Id != "" {
		query += fmt.Sprintf(" AND u.id=$%d", argIdx)
		args = append(args, filtro.Id)
		argIdx++
	}
	if filtro.ClienteId != "" {
		query += fmt.Sprintf(" AND uc.cliente_id=$%d", argIdx)
		args = append(args, filtro.ClienteId)
		argIdx++
	}
	if filtro.Nome != "" {
		query += fmt.Sprintf(" AND u.nome ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Nome+"%")
		argIdx++
	}
	if filtro.Email != "" {
		query += fmt.Sprintf(" AND u.email ILIKE $%d", argIdx)
		args = append(args, "%"+filtro.Email+"%")
		argIdx++
	}
	if filtro.Role != "" {
		query += fmt.Sprintf(" AND u.role=$%d", argIdx)
		args = append(args, filtro.Role)
		argIdx++
	}
	if filtro.Status != "" {
		query += fmt.Sprintf(" AND u.status=$%d", argIdx)
		args = append(args, filtro.Status)
		argIdx++
	}
	if filtro.DataInicio != "" {
		query += fmt.Sprintf(" AND u.criado_em >= $%d", argIdx)
		args = append(args, filtro.DataInicio)
		argIdx++
	}
	if filtro.DataFim != "" {
		query += fmt.Sprintf(" AND u.criado_em <= $%d", argIdx)
		args = append(args, filtro.DataFim)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM (" + query + ") t"
	if err := r.connection.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar usuarios: %w", err)
	}

	coluna := colunaOrdenacao(filtro.OrdenarPor, "criado_em", map[string]string{
		"nome": "u.nome", "email": "u.email", "role": "u.role", "status": "u.status", "criado_em": "u.criado_em", "atualizado_em": "u.atualizado_em",
	})
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", coluna, direcaoOrdenacao(filtro.Ordem), argIdx, argIdx+1)
	args = append(args, filtro.PorPagina, offset)
	rows, err := r.connection.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listar usuarios: %w", err)
	}
	defer rows.Close()

	var usuarios []*usuario.Usuario
	for rows.Next() {
		u, err := scanner.Usuario(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("escanear usuario: %w", err)
		}
		usuarios = append(usuarios, u)
	}
	return usuarios, total, nil
}

func (r *UsuarioRepository) SalvarAuth(ctx context.Context, auth *usuario.UsuarioAuth) error {
	auth.Id = uuid.New().String()
	stmts := statement.Get()
	err := r.connection.Exec(ctx, stmts.SalvarUsuarioAuth, pgx.NamedArgs{
		"id":             auth.Id,
		"usuario_id":     auth.UsuarioId,
		"tipo":           auth.Tipo,
		"provider_id":    auth.ProviderId,
		"provider_email": nilableStr(auth.ProviderEmail),
		"criado_em":      auth.CriadoEm,
	})
	if err != nil {
		return fmt.Errorf("salvar usuario auth: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) BuscarAuthPorProvider(ctx context.Context, tipo, providerID string) (*usuario.UsuarioAuth, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarUsuarioAuthPorProvider, pgx.NamedArgs{
		"tipo":        tipo,
		"provider_id": providerID,
	})
	auth, err := scanner.UsuarioAuth(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usuario.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar usuario auth por provider: %w", err)
	}
	return auth, nil
}

func (r *UsuarioRepository) SalvarRefreshToken(ctx context.Context, token *usuario.RefreshToken) error {
	token.Id = uuid.New().String()
	stmts := statement.Get()
	if err := r.connection.Exec(ctx, stmts.SalvarUsuarioRefreshToken, pgx.NamedArgs{
		"id":         token.Id,
		"usuario_id": token.UsuarioId,
		"token_hash": token.TokenHash,
		"expires_at": token.ExpiresAt,
		"criado_em":  token.CriadoEm,
	}); err != nil {
		return fmt.Errorf("salvar refresh token usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) BuscarRefreshTokenPorHash(ctx context.Context, tokenHash string) (*usuario.RefreshToken, error) {
	stmts := statement.Get()
	row := r.connection.QueryRow(ctx, stmts.BuscarUsuarioRefreshTokenPorHash, pgx.NamedArgs{"token_hash": tokenHash})

	token := &usuario.RefreshToken{}
	if err := row.Scan(&token.Id, &token.UsuarioId, &token.TokenHash, &token.ExpiresAt, &token.RevogadoEm, &token.CriadoEm); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usuario.ErrNaoAutorizado
		}
		return nil, fmt.Errorf("buscar refresh token usuario: %w", err)
	}
	return token, nil
}

func (r *UsuarioRepository) RevogarRefreshToken(ctx context.Context, tokenHash string, revogadoEm time.Time) error {
	if tokenHash == "" {
		return nil
	}
	stmts := statement.Get()
	if err := r.connection.Exec(ctx, stmts.RevogarUsuarioRefreshToken, pgx.NamedArgs{
		"token_hash":  tokenHash,
		"revogado_em": revogadoEm,
	}); err != nil {
		return fmt.Errorf("revogar refresh token usuario: %w", err)
	}
	return nil
}

func nilableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
