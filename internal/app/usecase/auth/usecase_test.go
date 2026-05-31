package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"nous/internal/app/dto"
	"nous/internal/domain/cliente"
	"nous/internal/domain/usuario"

	"golang.org/x/crypto/bcrypt"
)

type usuarioRepositoryFake struct {
	porEmail map[string]*usuario.Usuario
	porID    map[string]*usuario.Usuario
	refresh  map[string]*usuario.RefreshToken
	auth     *usuario.UsuarioAuth

	salvoRefresh    *usuario.RefreshToken
	revogadoHash    string
	revogadoUsuario string
	salvoUsuario    *usuario.Usuario
	salvoAuth       *usuario.UsuarioAuth
	salvoReset      *usuario.ResetSenhaToken
	resetTokens     map[string]*usuario.ResetSenhaToken
	resetUsado      string
}

func (r *usuarioRepositoryFake) Salvar(_ context.Context, u *usuario.Usuario) error {
	if u.Id == "" {
		u.Id = "usuario-novo"
	}
	r.salvoUsuario = u
	return nil
}

func (r *usuarioRepositoryFake) BuscarPorID(_ context.Context, id string) (*usuario.Usuario, error) {
	if u, ok := r.porID[id]; ok {
		return u, nil
	}
	return nil, usuario.ErrNaoEncontrado
}

func (r *usuarioRepositoryFake) BuscarPorEmail(_ context.Context, email string) (*usuario.Usuario, error) {
	if u, ok := r.porEmail[email]; ok {
		return u, nil
	}
	return nil, usuario.ErrNaoEncontrado
}

func (r *usuarioRepositoryFake) Atualizar(context.Context, *usuario.Usuario) error { return nil }
func (r *usuarioRepositoryFake) SoftDelete(context.Context, string) error          { return nil }

func (r *usuarioRepositoryFake) Listar(context.Context, usuario.FiltroListagem) ([]*usuario.Usuario, int, error) {
	return nil, 0, nil
}

func (r *usuarioRepositoryFake) SalvarAuth(_ context.Context, auth *usuario.UsuarioAuth) error {
	r.salvoAuth = auth
	return nil
}

func (r *usuarioRepositoryFake) BuscarAuthPorProvider(context.Context, string, string) (*usuario.UsuarioAuth, error) {
	if r.auth != nil {
		return r.auth, nil
	}
	return nil, usuario.ErrNaoEncontrado
}

func (r *usuarioRepositoryFake) SalvarRefreshToken(_ context.Context, token *usuario.RefreshToken) error {
	r.salvoRefresh = token
	return nil
}

func (r *usuarioRepositoryFake) BuscarRefreshTokenPorHash(_ context.Context, tokenHash string) (*usuario.RefreshToken, error) {
	if token, ok := r.refresh[tokenHash]; ok {
		return token, nil
	}
	return nil, usuario.ErrNaoAutorizado
}

func (r *usuarioRepositoryFake) RevogarRefreshToken(_ context.Context, tokenHash string, _ time.Time) error {
	r.revogadoHash = tokenHash
	return nil
}

func (r *usuarioRepositoryFake) RevogarRefreshTokensDoUsuario(_ context.Context, usuarioId string, _ time.Time) error {
	r.revogadoUsuario = usuarioId
	return nil
}

func (r *usuarioRepositoryFake) ExpurgarRefreshTokens(context.Context, time.Time) error { return nil }
func (r *usuarioRepositoryFake) SalvarResetSenhaToken(_ context.Context, token *usuario.ResetSenhaToken) error {
	r.salvoReset = token
	return nil
}
func (r *usuarioRepositoryFake) BuscarResetSenhaTokenPorHash(_ context.Context, tokenHash string) (*usuario.ResetSenhaToken, error) {
	if token, ok := r.resetTokens[tokenHash]; ok {
		return token, nil
	}
	return nil, usuario.ErrNaoEncontrado
}
func (r *usuarioRepositoryFake) MarcarResetSenhaTokenUsado(_ context.Context, tokenHash string, _ time.Time) error {
	r.resetUsado = tokenHash
	return nil
}

type clienteRepositoryFake struct {
	clientes []*cliente.Cliente
}

func (r *clienteRepositoryFake) Salvar(context.Context, *cliente.Cliente) error { return nil }
func (r *clienteRepositoryFake) BuscarPorID(context.Context, string) (*cliente.Cliente, error) {
	return nil, cliente.ErrNaoEncontrado
}
func (r *clienteRepositoryFake) Atualizar(context.Context, *cliente.Cliente) error { return nil }
func (r *clienteRepositoryFake) SoftDelete(context.Context, string) error          { return nil }
func (r *clienteRepositoryFake) Listar(context.Context, cliente.FiltroListagem) ([]*cliente.Cliente, int, error) {
	return nil, 0, nil
}
func (r *clienteRepositoryFake) VincularUsuario(context.Context, *cliente.UsuarioCliente) error {
	return nil
}
func (r *clienteRepositoryFake) ListarPorUsuario(context.Context, string) ([]*cliente.Cliente, error) {
	return r.clientes, nil
}
func (r *clienteRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return nil
}

type tokenServiceFake struct {
	accessToken   string
	refreshToken  string
	refreshUserID string
	gerarChamadas int
}

func (s *tokenServiceFake) GerarTokens(string, string, string) (string, string, error) {
	s.gerarChamadas++
	return s.accessToken, s.refreshToken, nil
}

func (s *tokenServiceFake) ValidarRefreshToken(string) (string, error) {
	if s.refreshUserID == "" {
		return "", errors.New("refresh invalido")
	}
	return s.refreshUserID, nil
}

func TestLoginSalvaRefreshHashEListaClientes(t *testing.T) {
	senhaHash, err := bcrypt.GenerateFromPassword([]byte("senha-segura"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("gerar hash: %v", err)
	}
	repository := &usuarioRepositoryFake{porEmail: map[string]*usuario.Usuario{
		"user@sognix.com.br": {
			Id:        "usuario-1",
			Nome:      "Usuario Teste",
			Email:     "user@sognix.com.br",
			SenhaHash: string(senhaHash),
			Role:      usuario.RoleUsuario,
			Status:    usuario.StatusAtivo,
		},
	}}
	tokenService := &tokenServiceFake{accessToken: "access-token", refreshToken: "refresh-token"}
	useCase := NewUseCase(repository, &clienteRepositoryFake{clientes: []*cliente.Cliente{{Id: "cliente-1", Nome: "Cliente", Status: cliente.StatusAtivo}}}, tokenService)

	output, err := useCase.Login(context.Background(), dto.LoginInput{Email: "user@sognix.com.br", Senha: "senha-segura"})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Token != "access-token" || output.RefreshToken != "refresh-token" || len(output.Clientes) != 1 {
		t.Fatalf("saida de login inesperada: %+v", output)
	}
	if repository.salvoRefresh == nil || repository.salvoRefresh.TokenHash != hashToken("refresh-token") {
		t.Fatal("refresh token deveria ser salvo apenas como hash")
	}
	if repository.salvoRefresh.TokenHash == "refresh-token" {
		t.Fatal("refresh token em texto puro nao deveria ser persistido")
	}
}

func TestLoginRecusaUsuarioInativoSemGerarToken(t *testing.T) {
	repository := &usuarioRepositoryFake{porEmail: map[string]*usuario.Usuario{
		"user@sognix.com.br": {Id: "usuario-1", Email: "user@sognix.com.br", Status: usuario.StatusInativo},
	}}
	tokenService := &tokenServiceFake{accessToken: "access-token", refreshToken: "refresh-token"}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, tokenService)

	_, err := useCase.Login(context.Background(), dto.LoginInput{Email: "user@sognix.com.br", Senha: "senha"})

	if !errors.Is(err, usuario.ErrNaoAutorizado) {
		t.Fatalf("erro esperado ErrNaoAutorizado, recebido %v", err)
	}
	if tokenService.gerarChamadas != 0 || repository.salvoRefresh != nil {
		t.Fatal("usuario inativo nao deveria gerar ou persistir tokens")
	}
}

func TestRefreshTokenRotacionaERevogaAnterior(t *testing.T) {
	repository := &usuarioRepositoryFake{
		porID: map[string]*usuario.Usuario{
			"usuario-1": {Id: "usuario-1", Nome: "Usuario", Email: "user@sognix.com.br", Role: usuario.RoleUsuario, Status: usuario.StatusAtivo},
		},
		refresh: map[string]*usuario.RefreshToken{
			hashToken("refresh-antigo"): {UsuarioId: "usuario-1", TokenHash: hashToken("refresh-antigo"), ExpiresAt: time.Now().Add(time.Hour)},
		},
	}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &tokenServiceFake{accessToken: "access-novo", refreshToken: "refresh-novo", refreshUserID: "usuario-1"})

	output, err := useCase.RefreshToken(context.Background(), dto.RefreshTokenInput{RefreshToken: "refresh-antigo"})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.RefreshToken != "refresh-novo" {
		t.Fatalf("refresh novo esperado, recebido %s", output.RefreshToken)
	}
	if repository.revogadoHash != hashToken("refresh-antigo") {
		t.Fatal("refresh token anterior deveria ser revogado")
	}
	if repository.salvoRefresh == nil || repository.salvoRefresh.TokenHash != hashToken("refresh-novo") {
		t.Fatal("novo refresh token deveria ser persistido como hash")
	}
}

func TestLoginGoogleCriaUsuarioEAuthQuandoProviderNovo(t *testing.T) {
	repository := &usuarioRepositoryFake{}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &tokenServiceFake{accessToken: "access", refreshToken: "refresh"})

	_, err := useCase.LoginGoogle(context.Background(), GoogleCallbackInput{ProviderId: "google-1", ProviderEmail: "user@sognix.com.br", Nome: "Usuario Google"})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repository.salvoUsuario == nil || repository.salvoUsuario.Email != "user@sognix.com.br" || repository.salvoUsuario.Status != usuario.StatusAtivo {
		t.Fatalf("usuario google nao foi criado corretamente: %+v", repository.salvoUsuario)
	}
	if repository.salvoAuth == nil || repository.salvoAuth.ProviderId != "google-1" || repository.salvoAuth.UsuarioId != repository.salvoUsuario.Id {
		t.Fatalf("auth google nao foi vinculada corretamente: %+v", repository.salvoAuth)
	}
}

func TestEsqueciSenhaSalvaTokenHasheadoSemEnumerarEmail(t *testing.T) {
	t.Setenv("RESET_PASSWORD_RETURN_TOKEN", "true")
	repository := &usuarioRepositoryFake{porEmail: map[string]*usuario.Usuario{
		"user@sognix.com.br": {Id: "usuario-1", Email: "user@sognix.com.br", Status: usuario.StatusAtivo},
	}}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &tokenServiceFake{})

	output, err := useCase.EsqueciSenha(context.Background(), dto.EsqueciSenhaInput{Email: "user@sognix.com.br"})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.ResetToken == "" || repository.salvoReset == nil {
		t.Fatalf("token de reset esperado em ambiente de teste: output=%+v salvo=%+v", output, repository.salvoReset)
	}
	if repository.salvoReset.TokenHash == output.ResetToken {
		t.Fatal("token de reset nao deve ser persistido em texto puro")
	}
	if repository.salvoReset.TokenHash != hashToken(output.ResetToken) {
		t.Fatal("hash persistido nao corresponde ao token retornado")
	}

	output, err = useCase.EsqueciSenha(context.Background(), dto.EsqueciSenhaInput{Email: "ausente@sognix.com.br"})
	if err != nil || output.Mensagem == "" {
		t.Fatalf("email ausente deveria retornar mensagem generica sem erro: output=%+v err=%v", output, err)
	}
}

func TestRedefinirSenhaUsaTokenUnico(t *testing.T) {
	token := "reset-token"
	senhaAntiga, err := bcrypt.GenerateFromPassword([]byte("antiga"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("gerar hash: %v", err)
	}
	repository := &usuarioRepositoryFake{
		porID: map[string]*usuario.Usuario{
			"usuario-1": {Id: "usuario-1", SenhaHash: string(senhaAntiga), Status: usuario.StatusAtivo, Version: 2},
		},
		resetTokens: map[string]*usuario.ResetSenhaToken{
			hashToken(token): {UsuarioId: "usuario-1", TokenHash: hashToken(token), ExpiresAt: time.Now().Add(time.Minute)},
		},
	}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &tokenServiceFake{})

	if err := useCase.RedefinirSenha(context.Background(), dto.RedefinirSenhaInput{Token: token, NovaSenha: "nova"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repository.resetUsado != hashToken(token) {
		t.Fatal("token deveria ser marcado como usado")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.porID["usuario-1"].SenhaHash), []byte("nova")); err != nil {
		t.Fatalf("senha nova nao foi persistida: %v", err)
	}
	if repository.revogadoUsuario != "usuario-1" {
		t.Fatalf("refresh tokens do usuario deveriam ser revogados, recebido %q", repository.revogadoUsuario)
	}
}
