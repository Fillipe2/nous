package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/usuario"
)

// Usuario escaneia uma row do banco para a entidade de domínio Usuario.
func Usuario(row pgx.Row) (*usuario.Usuario, error) {
	var (
		id            string
		nome          string
		email         string
		senhaHash     *string
		role          string
		status        string
		criadoEm      time.Time
		atualizadoEm  time.Time
		version       int
		emailVerified bool
		ultimoLoginEm *time.Time
	)
	if err := row.Scan(
		&id, &nome, &email, &senhaHash,
		&role, &status, &criadoEm, &atualizadoEm, &version,
		&emailVerified, &ultimoLoginEm,
	); err != nil {
		return nil, err
	}
	u := &usuario.Usuario{
		Id:            id,
		Nome:          nome,
		Email:         email,
		Role:          usuario.Role(role),
		Status:        usuario.Status(status),
		EmailVerified: emailVerified,
		UltimoLoginEm: ultimoLoginEm,
		CriadoEm:      criadoEm,
		AtualizadoEm:  atualizadoEm,
		Version:       version,
	}
	if senhaHash != nil {
		u.SenhaHash = *senhaHash
	}
	return u, nil
}

// UsuarioAuth escaneia uma row do banco para UsuarioAuth.
func UsuarioAuth(row pgx.Row) (*usuario.UsuarioAuth, error) {
	var (
		id            string
		usuarioID     string
		tipo          string
		providerID    string
		providerEmail *string
		criadoEm      time.Time
	)
	if err := row.Scan(&id, &usuarioID, &tipo, &providerID, &providerEmail, &criadoEm); err != nil {
		return nil, err
	}
	auth := &usuario.UsuarioAuth{
		Id:         id,
		UsuarioId:  usuarioID,
		Tipo:       tipo,
		ProviderId: providerID,
		CriadoEm:   criadoEm,
	}
	if providerEmail != nil {
		auth.ProviderEmail = *providerEmail
	}
	return auth, nil
}

func ResetSenhaToken(row pgx.Row) (*usuario.ResetSenhaToken, error) {
	var token usuario.ResetSenhaToken
	if err := row.Scan(
		&token.Id,
		&token.UsuarioId,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsadoEm,
		&token.CriadoEm,
	); err != nil {
		return nil, err
	}
	return &token, nil
}

func VerificacaoEmailToken(row pgx.Row) (*usuario.VerificacaoEmailToken, error) {
	var token usuario.VerificacaoEmailToken
	if err := row.Scan(
		&token.Id,
		&token.UsuarioId,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsadoEm,
		&token.CriadoEm,
	); err != nil {
		return nil, err
	}
	return &token, nil
}
