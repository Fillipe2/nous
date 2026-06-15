package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"nous/internal/app/dto"
	"nous/internal/domain/negocio"
	"nous/internal/domain/usuario"

	"golang.org/x/crypto/bcrypt"
)

const duracaoRefreshToken = 7 * 24 * time.Hour
const duracaoResetSenhaToken = 30 * time.Minute
const duracaoVerificacaoEmailToken = 24 * time.Hour
const duracaoMinimaEsqueciSenha = 150 * time.Millisecond

// hashSenhaDummy é um hash bcrypt válido usado para igualar o tempo de resposta
// do login quando o e-mail não existe ou o usuário está inativo, evitando
// enumeração de usuários por timing.
var hashSenhaDummy, _ = bcrypt.GenerateFromPassword([]byte("dummy-timing-equalization-secret"), bcrypt.DefaultCost)

type TokenService interface {
	GerarTokens(usuarioId, email, role string) (accessToken, refreshToken string, err error)
	ValidarRefreshToken(token string) (usuarioId string, err error)
}

type UseCase struct {
	repository        usuario.Repository
	negocioRepository negocio.Repository
	tokenService      TokenService
	loginThrottle     *loginThrottle
}

func NewUseCase(repository usuario.Repository, negocioRepository negocio.Repository, tokenService TokenService) *UseCase {
	return &UseCase{
		repository:        repository,
		negocioRepository: negocioRepository,
		tokenService:      tokenService,
		loginThrottle:     novoLoginThrottle(maxTentativasLogin, janelaTentativasLogin),
	}
}

func (u *UseCase) Login(ctx context.Context, input dto.LoginInput) (*dto.LoginOutput, error) {
	// Throttle por conta complementa o rate limit por IP do middleware. A
	// resposta é a mesma de credencial inválida para não revelar o bloqueio,
	// e o hash dummy mantém o tempo de resposta constante.
	if u.loginThrottle.Bloqueado(input.Email) {
		_ = bcrypt.CompareHashAndPassword(hashSenhaDummy, []byte(input.Senha))
		return nil, fmt.Errorf("login bloqueado temporariamente: %w", usuario.ErrNaoAutorizado)
	}

	us, err := u.repository.BuscarPorEmail(ctx, input.Email)
	if err != nil {
		// Compara contra um hash dummy para manter o tempo de resposta constante
		// e impedir enumeração de usuários por timing.
		_ = bcrypt.CompareHashAndPassword(hashSenhaDummy, []byte(input.Senha))
		u.loginThrottle.RegistrarFalha(input.Email)
		return nil, fmt.Errorf("login: %w", usuario.ErrNaoAutorizado)
	}
	if us.Status == usuario.StatusExcluido || us.Status == usuario.StatusInativo {
		_ = bcrypt.CompareHashAndPassword(hashSenhaDummy, []byte(input.Senha))
		u.loginThrottle.RegistrarFalha(input.Email)
		return nil, usuario.ErrNaoAutorizado
	}
	if err := bcrypt.CompareHashAndPassword([]byte(us.SenhaHash), []byte(input.Senha)); err != nil {
		u.loginThrottle.RegistrarFalha(input.Email)
		return nil, usuario.ErrNaoAutorizado
	}

	u.loginThrottle.RegistrarSucesso(input.Email)
	u.registrarUltimoLogin(ctx, us.Id)

	accessToken, refreshToken, err := u.gerarTokens(ctx, us)
	if err != nil {
		return nil, err
	}

	return u.montarLoginOutput(ctx, us, accessToken, refreshToken)
}

// registrarUltimoLogin grava o instante do login. É telemetria best-effort: uma
// falha aqui não impede o login (a escrita do refresh token logo a seguir já
// detectaria indisponibilidade do banco).
func (u *UseCase) registrarUltimoLogin(ctx context.Context, usuarioId string) {
	_ = u.repository.AtualizarUltimoLogin(ctx, usuarioId, time.Now())
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
		// Verifica se já existe um usuário com o mesmo email para vincular a conta.
		// O handler garante que email_verified=true antes de chegar aqui.
		existente, emailErr := u.repository.BuscarPorEmail(ctx, input.ProviderEmail)
		if emailErr == nil && existente != nil {
			usuarioAuth := &usuario.UsuarioAuth{
				UsuarioId:     existente.Id,
				Tipo:          "google",
				ProviderId:    input.ProviderId,
				ProviderEmail: input.ProviderEmail,
				CriadoEm:      time.Now(),
			}
			if salvarErr := u.repository.SalvarAuth(ctx, usuarioAuth); salvarErr != nil {
				return nil, fmt.Errorf("login google vincular conta: %w", salvarErr)
			}
			us = existente
		} else {
			us = &usuario.Usuario{
				Nome:          input.Nome,
				Email:         input.ProviderEmail,
				Role:          usuario.RoleUsuario,
				Status:        usuario.StatusAtivo,
				EmailVerified: true,
				CriadoEm:      time.Now(),
				AtualizadoEm:  time.Now(),
				Version:       1,
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

	u.registrarUltimoLogin(ctx, us.Id)

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

func (u *UseCase) EsqueciSenha(ctx context.Context, input dto.EsqueciSenhaInput) (*dto.EsqueciSenhaOutput, error) {
	inicio := time.Now()
	defer aguardarDuracaoMinima(inicio, duracaoMinimaEsqueciSenha)

	mensagem := "se o email estiver cadastrado, enviaremos as instrucoes para redefinir a senha"
	output := &dto.EsqueciSenhaOutput{Mensagem: mensagem}

	us, err := u.repository.BuscarPorEmail(ctx, strings.TrimSpace(input.Email))
	if err != nil || us.Status != usuario.StatusAtivo {
		return output, nil
	}

	token, err := gerarTokenSeguro()
	if err != nil {
		return nil, fmt.Errorf("gerar token reset senha: %w", err)
	}
	agora := time.Now()
	if err := u.repository.SalvarResetSenhaToken(ctx, &usuario.ResetSenhaToken{
		UsuarioId: us.Id,
		TokenHash: hashToken(token),
		ExpiresAt: agora.Add(duracaoResetSenhaToken),
		CriadoEm:  agora,
	}); err != nil {
		return nil, fmt.Errorf("salvar token reset senha: %w", err)
	}

	if resetPasswordReturnTokenHabilitado() {
		output.ResetToken = token
	}
	return output, nil
}

func (u *UseCase) RedefinirSenha(ctx context.Context, input dto.RedefinirSenhaInput) error {
	if strings.TrimSpace(input.Token) == "" || input.NovaSenha == "" {
		return usuario.ErrNaoAutorizado
	}
	if err := usuario.ValidarSenha(input.NovaSenha); err != nil {
		return err
	}

	resetToken, err := u.repository.BuscarResetSenhaTokenPorHash(ctx, hashToken(input.Token))
	if err != nil || resetToken.UsadoEm != nil || time.Now().After(resetToken.ExpiresAt) {
		return usuario.ErrNaoAutorizado
	}

	us, err := u.repository.BuscarPorID(ctx, resetToken.UsuarioId)
	if err != nil || us.Status != usuario.StatusAtivo {
		return usuario.ErrNaoAutorizado
	}
	if err := u.repository.MarcarResetSenhaTokenUsado(ctx, resetToken.TokenHash, time.Now()); err != nil {
		return fmt.Errorf("marcar token reset senha usado: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NovaSenha), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("redefinir senha gerar hash: %w", err)
	}
	us.SenhaHash = string(hash)
	us.AtualizadoEm = time.Now()
	us.Version++
	if err := u.repository.Atualizar(ctx, us); err != nil {
		return fmt.Errorf("redefinir senha salvar usuario: %w", err)
	}
	if err := u.repository.RevogarRefreshTokensDoUsuario(ctx, us.Id, time.Now()); err != nil {
		return fmt.Errorf("redefinir senha revogar sessoes: %w", err)
	}
	return nil
}

// SolicitarVerificacaoEmail gera um token de verificação para um usuário ativo
// e ainda não verificado. A resposta é neutra para não revelar cadastro.
func (u *UseCase) SolicitarVerificacaoEmail(ctx context.Context, input dto.SolicitarVerificacaoEmailInput) (*dto.SolicitarVerificacaoEmailOutput, error) {
	output := &dto.SolicitarVerificacaoEmailOutput{
		Mensagem: "se o email estiver cadastrado e ainda nao verificado, enviaremos as instrucoes de verificacao",
	}

	us, err := u.repository.BuscarPorEmail(ctx, strings.TrimSpace(input.Email))
	if err != nil || us.Status != usuario.StatusAtivo || us.EmailVerified {
		return output, nil
	}

	token, err := gerarTokenSeguro()
	if err != nil {
		return nil, fmt.Errorf("gerar token verificacao email: %w", err)
	}
	agora := time.Now()
	if err := u.repository.SalvarVerificacaoEmailToken(ctx, &usuario.VerificacaoEmailToken{
		UsuarioId: us.Id,
		TokenHash: hashToken(token),
		ExpiresAt: agora.Add(duracaoVerificacaoEmailToken),
		CriadoEm:  agora,
	}); err != nil {
		return nil, fmt.Errorf("salvar token verificacao email: %w", err)
	}

	if emailVerificationReturnTokenHabilitado() {
		output.VerificacaoToken = token
	}
	return output, nil
}

// VerificarEmail consome um token de verificação e marca o email do usuário como
// verificado. Token inválido, usado ou expirado retorna ErrNaoAutorizado.
func (u *UseCase) VerificarEmail(ctx context.Context, input dto.VerificarEmailInput) error {
	if strings.TrimSpace(input.Token) == "" {
		return usuario.ErrNaoAutorizado
	}

	token, err := u.repository.BuscarVerificacaoEmailTokenPorHash(ctx, hashToken(input.Token))
	if err != nil || token.UsadoEm != nil || time.Now().After(token.ExpiresAt) {
		return usuario.ErrNaoAutorizado
	}

	us, err := u.repository.BuscarPorID(ctx, token.UsuarioId)
	if err != nil || us.Status != usuario.StatusAtivo {
		return usuario.ErrNaoAutorizado
	}
	if err := u.repository.MarcarVerificacaoEmailTokenUsado(ctx, token.TokenHash, time.Now()); err != nil {
		return fmt.Errorf("marcar token verificacao email usado: %w", err)
	}

	if us.EmailVerified {
		return nil
	}
	us.EmailVerified = true
	us.AtualizadoEm = time.Now()
	us.Version++
	if err := u.repository.Atualizar(ctx, us); err != nil {
		return fmt.Errorf("verificar email salvar usuario: %w", err)
	}
	return nil
}

func (u *UseCase) montarLoginOutput(ctx context.Context, us *usuario.Usuario, accessToken, refreshToken string) (*dto.LoginOutput, error) {
	negocios, err := u.negocioRepository.ListarPorUsuario(ctx, us.Id)
	if err != nil {
		return nil, fmt.Errorf("listar negocios usuario login: %w", err)
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

	for _, c := range negocios {
		output.Negocios = append(output.Negocios, &dto.NegocioBasicoOutput{
			Id:       c.Id,
			Nome:     c.Nome,
			Email:    c.Email,
			Telefone: c.Telefone,
			Whatsapp: c.Whatsapp,
			Status:   string(c.Status),
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

func gerarTokenSeguro() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func resetPasswordReturnTokenHabilitado() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("RESET_PASSWORD_RETURN_TOKEN")), "true")
}

func emailVerificationReturnTokenHabilitado() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("EMAIL_VERIFICATION_RETURN_TOKEN")), "true")
}

func aguardarDuracaoMinima(inicio time.Time, duracao time.Duration) {
	if restante := duracao - time.Since(inicio); restante > 0 {
		time.Sleep(restante)
	}
}
