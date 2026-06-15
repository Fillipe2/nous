package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/usuario"
	"nous/internal/infra/database/postgres/internal/scanner"
	"nous/internal/infra/database/postgres/internal/statement"

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
		"id":             u.Id,
		"nome":           u.Nome,
		"email":          u.Email,
		"senha_hash":     nilableStr(u.SenhaHash),
		"role":           string(u.Role),
		"status":         string(u.Status),
		"email_verified": u.EmailVerified,
		"criado_em":      u.CriadoEm,
		"atualizado_em":  u.AtualizadoEm,
		"version":        u.Version,
	})
	if err != nil {
		if isUniqueViolation(err, "usuario_email") {
			return usuario.ErrEmailDuplicado
		}
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
		"email_verified":   u.EmailVerified,
		"atualizado_em":    u.AtualizadoEm,
		"version":          u.Version,
		"version_anterior": u.Version - 1,
		"id":               u.Id,
	})
	if err != nil {
		if isUniqueViolation(err, "usuario_email") {
			return usuario.ErrEmailDuplicado
		}
		return fmt.Errorf("atualizar usuario: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return usuario.ErrVersaoConflito
	}
	return nil
}

func (r *UsuarioRepository) AtualizarUltimoLogin(ctx context.Context, id string, em time.Time) error {
	stmts := statement.Get()
	if err := r.connection.Exec(ctx, stmts.AtualizarUltimoLoginUsuario, pgx.NamedArgs{
		"id":              id,
		"ultimo_login_em": em,
	}); err != nil {
		return fmt.Errorf("atualizar ultimo login usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) ContarAdministradoresAtivos(ctx context.Context) (int, error) {
	stmts := statement.Get()
	var total int
	if err := r.connection.QueryRow(ctx, stmts.ContarAdministradoresAtivos).Scan(&total); err != nil {
		return 0, fmt.Errorf("contar administradores ativos: %w", err)
	}
	return total, nil
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
	query := `SELECT u.id, u.nome, u.email, u.senha_hash, u.role, u.status, u.criado_em, u.atualizado_em, u.version, u.email_verified, u.ultimo_login_em FROM usuario u`
	args := []any{}
	argIdx := 1
	if filtro.NegocioId != "" {
		query += " INNER JOIN usuario_negocio uc ON uc.usuario_id=u.id"
	}
	query += " WHERE u.status != 'excluido'"
	if filtro.Id != "" {
		query += fmt.Sprintf(" AND u.id=$%d", argIdx)
		args = append(args, filtro.Id)
		argIdx++
	}
	if filtro.NegocioId != "" {
		query += fmt.Sprintf(" AND uc.negocio_id=$%d", argIdx)
		args = append(args, filtro.NegocioId)
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
		args = append(args, dataInicioFiltro(filtro.DataInicio))
		argIdx++
	}
	if filtro.DataFim != "" {
		operador, valor := dataFimFiltro(filtro.DataFim)
		query += fmt.Sprintf(" AND u.criado_em %s $%d", operador, argIdx)
		args = append(args, valor)
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

func (r *UsuarioRepository) SalvarResetSenhaToken(ctx context.Context, token *usuario.ResetSenhaToken) error {
	token.Id = uuid.New().String()
	st := statement.Get()
	err := r.connection.Exec(ctx, st.SalvarUsuarioResetSenhaToken, pgx.NamedArgs{
		"id":         token.Id,
		"usuario_id": token.UsuarioId,
		"token_hash": token.TokenHash,
		"expires_at": token.ExpiresAt,
		"criado_em":  token.CriadoEm,
	})
	if err != nil {
		return fmt.Errorf("salvar reset senha token: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) BuscarResetSenhaTokenPorHash(ctx context.Context, tokenHash string) (*usuario.ResetSenhaToken, error) {
	st := statement.Get()
	row := r.connection.QueryRow(ctx, st.BuscarUsuarioResetSenhaTokenPorHash, pgx.NamedArgs{"token_hash": tokenHash})
	token, err := scanner.ResetSenhaToken(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usuario.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar reset senha token: %w", err)
	}
	return token, nil
}

func (r *UsuarioRepository) MarcarResetSenhaTokenUsado(ctx context.Context, tokenHash string, usadoEm time.Time) error {
	st := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, st.MarcarUsuarioResetSenhaTokenUsado, pgx.NamedArgs{
		"token_hash": tokenHash,
		"usado_em":   usadoEm,
	})
	if err != nil {
		return fmt.Errorf("marcar reset senha token usado: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return usuario.ErrNaoEncontrado
	}
	return nil
}

func (r *UsuarioRepository) SalvarVerificacaoEmailToken(ctx context.Context, token *usuario.VerificacaoEmailToken) error {
	token.Id = uuid.New().String()
	st := statement.Get()
	err := r.connection.Exec(ctx, st.SalvarUsuarioVerificacaoEmailToken, pgx.NamedArgs{
		"id":         token.Id,
		"usuario_id": token.UsuarioId,
		"token_hash": token.TokenHash,
		"expires_at": token.ExpiresAt,
		"criado_em":  token.CriadoEm,
	})
	if err != nil {
		return fmt.Errorf("salvar verificacao email token: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) BuscarVerificacaoEmailTokenPorHash(ctx context.Context, tokenHash string) (*usuario.VerificacaoEmailToken, error) {
	st := statement.Get()
	row := r.connection.QueryRow(ctx, st.BuscarUsuarioVerificacaoEmailTokenPorHash, pgx.NamedArgs{"token_hash": tokenHash})
	token, err := scanner.VerificacaoEmailToken(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usuario.ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar verificacao email token: %w", err)
	}
	return token, nil
}

func (r *UsuarioRepository) MarcarVerificacaoEmailTokenUsado(ctx context.Context, tokenHash string, usadoEm time.Time) error {
	st := statement.Get()
	commandTag, err := r.connection.ExecResult(ctx, st.MarcarUsuarioVerificacaoEmailTokenUsado, pgx.NamedArgs{
		"token_hash": tokenHash,
		"usado_em":   usadoEm,
	})
	if err != nil {
		return fmt.Errorf("marcar verificacao email token usado: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return usuario.ErrNaoEncontrado
	}
	return nil
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

func (r *UsuarioRepository) RevogarRefreshTokensDoUsuario(ctx context.Context, usuarioId string, revogadoEm time.Time) error {
	if usuarioId == "" {
		return nil
	}
	stmts := statement.Get()
	if err := r.connection.Exec(ctx, stmts.RevogarUsuarioRefreshTokensDoUsuario, pgx.NamedArgs{
		"usuario_id":  usuarioId,
		"revogado_em": revogadoEm,
	}); err != nil {
		return fmt.Errorf("revogar refresh tokens do usuario: %w", err)
	}
	return nil
}

func (r *UsuarioRepository) ExpurgarRefreshTokens(ctx context.Context, antesDe time.Time) error {
	stmts := statement.Get()
	if err := r.connection.Exec(ctx, stmts.ExpurgarUsuarioRefreshTokens, pgx.NamedArgs{"antes_de": antesDe}); err != nil {
		return fmt.Errorf("expurgar refresh tokens usuario: %w", err)
	}
	return nil
}

func nilableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
