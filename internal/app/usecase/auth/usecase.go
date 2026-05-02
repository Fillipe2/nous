package auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/usuario"

	"golang.org/x/crypto/bcrypt"
)

const duracaoRefreshToken = 7 * 24 * time.Hour

type TokenService interface {
	GerarTokens(usuarioId, email, role string) (accessToken, refreshToken string, err error)
	ValidarRefreshToken(token string) (usuarioId string, err error)
}

type UseCase struct {
	repository        usuario.Repository
	clienteRepository cliente.Repository
	tokenService      TokenService
}

func NewUseCase(repository usuario.Repository, clienteRepository cliente.Repository, tokenService TokenService) *UseCase {
	return &UseCase{repository: repository, clienteRepository: clienteRepository, tokenService: tokenService}
}

func (u *UseCase) Login(ctx context.Context, input dto.LoginInput) (*dto.LoginOutput, error) {
	us, err := u.repository.BuscarPorEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("login: %w", usuario.ErrNaoAutorizado)
	}
	if us.Status == usuario.StatusExcluido || us.Status == usuario.StatusInativo {
		return nil, usuario.ErrNaoAutorizado
	}
	if err := bcrypt.CompareHashAndPassword([]byte(us.SenhaHash), []byte(input.Senha)); err != nil {
		return nil, usuario.ErrNaoAutorizado
	}

	accessToken, refreshToken, err := u.gerarTokens(ctx, us)
	if err != nil {
		return nil, err
	}

	return u.montarLoginOutput(ctx, us, accessToken, refreshToken)
}

func (u *UseCase) RefreshToken(ctx context.Context, input dto.RefreshTokenInput) (*dto.LoginOutput, error) {
	usuarioId, err := u.tokenService.ValidarRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token invalido: %w", usuario.ErrNaoAutorizado)
	}
	refreshPersistido, err := u.repository.BuscarRefreshTokenPorHash(ctx, hashToken(input.RefreshToken))
	if err != nil || refreshPersistido.UsuarioId != usuarioId || refreshPersistido.RevogadoEm != nil || time.Now().After(refreshPersistido.ExpiresAt) {
		return nil, fmt.Errorf("refresh token revogado: %w", usuario.ErrNaoAutorizado)
	}

	us, err := u.repository.BuscarPorID(ctx, usuarioId)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	if us.Status != usuario.StatusAtivo {
		return nil, usuario.ErrNaoAutorizado
	}

	accessToken, refreshToken, err := u.gerarTokens(ctx, us)
	if err != nil {
		return nil, err
	}
	if err := u.repository.RevogarRefreshToken(ctx, refreshPersistido.TokenHash, time.Now()); err != nil {
		return nil, fmt.Errorf("revogar refresh token anterior: %w", err)
	}

	return u.montarLoginOutput(ctx, us, accessToken, refreshToken)
}

// GoogleCallbackInput contém os dados retornados pelo provider Google.
type GoogleCallbackInput struct {
	ProviderId    string
	ProviderEmail string
	Nome          string
}

func (u *UseCase) LoginGoogle(ctx context.Context, input GoogleCallbackInput) (*dto.LoginOutput, error) {
	auth, err := u.repository.BuscarAuthPorProvider(ctx, "google", input.ProviderId)
	var us *usuario.Usuario

	if err != nil {
		us = &usuario.Usuario{
			Nome:         input.Nome,
			Email:        input.ProviderEmail,
			Role:         usuario.RoleUsuario,
			Status:       usuario.StatusAtivo,
			CriadoEm:     time.Now(),
			AtualizadoEm: time.Now(),
			Version:      1,
		}
		if salvarErr := u.repository.Salvar(ctx, us); salvarErr != nil {
			return nil, fmt.Errorf("login google criar usuario: %w", salvarErr)
		}
		usuarioAuth := &usuario.UsuarioAuth{
			UsuarioId:     us.Id,
			Tipo:          "google",
			ProviderId:    input.ProviderId,
			ProviderEmail: input.ProviderEmail,
			CriadoEm:      time.Now(),
		}
		if salvarErr := u.repository.SalvarAuth(ctx, usuarioAuth); salvarErr != nil {
			return nil, fmt.Errorf("login google salvar auth: %w", salvarErr)
		}
	} else {
		us, err = u.repository.BuscarPorID(ctx, auth.UsuarioId)
		if err != nil {
			return nil, fmt.Errorf("login google buscar usuario: %w", err)
		}
	}

	if us.Status != usuario.StatusAtivo {
		return nil, usuario.ErrNaoAutorizado
	}

	accessToken, refreshToken, err := u.gerarTokens(ctx, us)
	if err != nil {
		return nil, err
	}

	return u.montarLoginOutput(ctx, us, accessToken, refreshToken)
}

func (u *UseCase) Logout(ctx context.Context, input dto.LogoutInput) error {
	if input.RefreshToken == "" {
		return nil
	}
	if err := u.repository.RevogarRefreshToken(ctx, hashToken(input.RefreshToken), time.Now()); err != nil {
		return fmt.Errorf("logout revogar refresh token: %w", err)
	}
	return nil
}

func (u *UseCase) montarLoginOutput(ctx context.Context, us *usuario.Usuario, accessToken, refreshToken string) (*dto.LoginOutput, error) {
	clientes, err := u.clienteRepository.ListarPorUsuario(ctx, us.Id)
	if err != nil {
		return nil, fmt.Errorf("listar clientes usuario login: %w", err)
	}

	output := &dto.LoginOutput{
		Token:        accessToken,
		RefreshToken: refreshToken,
		Usuario: &dto.UsuarioOutput{
			Id:           us.Id,
			Nome:         us.Nome,
			Email:        us.Email,
			Role:         string(us.Role),
			Status:       string(us.Status),
			CriadoEm:     us.CriadoEm,
			AtualizadoEm: us.AtualizadoEm,
		},
	}

	for _, c := range clientes {
		output.Clientes = append(output.Clientes, &dto.ClienteBasicoOutput{
			Id:     c.Id,
			Nome:   c.Nome,
			Status: string(c.Status),
		})
	}

	return output, nil
}

func (u *UseCase) gerarTokens(ctx context.Context, us *usuario.Usuario) (accessToken, refreshToken string, err error) {
	accessToken, refreshToken, err = u.tokenService.GerarTokens(us.Id, us.Email, string(us.Role))
	if err != nil {
		return "", "", fmt.Errorf("gerar tokens: %w", err)
	}

	agora := time.Now()
	if err := u.repository.SalvarRefreshToken(ctx, &usuario.RefreshToken{
		UsuarioId: us.Id,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: agora.Add(duracaoRefreshToken),
		CriadoEm:  agora,
	}); err != nil {
		return "", "", fmt.Errorf("salvar refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}
