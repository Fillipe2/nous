package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nous/internal/app/dto"
	uccliente "nous/internal/app/usecase/cliente"
	ucdashboard "nous/internal/app/usecase/dashboard"
	uclead "nous/internal/app/usecase/lead"
	ucusuario "nous/internal/app/usecase/usuario"
	"nous/internal/domain/cliente"
	"nous/internal/domain/dashboard"
	"nous/internal/domain/evento"
	"nous/internal/domain/landingpage"
	"nous/internal/domain/lead"
	"nous/internal/domain/usuario"
	infrajwt "nous/internal/infra/jwt"
	"nous/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type usuarioHandlerRepositoryFake struct {
	usuario    *usuario.Usuario
	atualizado *usuario.Usuario
	salvo      *usuario.Usuario
}

func (r *usuarioHandlerRepositoryFake) Salvar(_ context.Context, u *usuario.Usuario) error {
	u.Id = "usuario-http"
	r.salvo = u
	return nil
}
func (r *usuarioHandlerRepositoryFake) BuscarPorID(context.Context, string) (*usuario.Usuario, error) {
	if r.usuario == nil {
		return nil, usuario.ErrNaoEncontrado
	}
	return r.usuario, nil
}
func (r *usuarioHandlerRepositoryFake) BuscarPorEmail(context.Context, string) (*usuario.Usuario, error) {
	return nil, usuario.ErrNaoEncontrado
}
func (r *usuarioHandlerRepositoryFake) Atualizar(_ context.Context, u *usuario.Usuario) error {
	r.atualizado = u
	return nil
}
func (r *usuarioHandlerRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *usuarioHandlerRepositoryFake) Listar(context.Context, usuario.FiltroListagem) ([]*usuario.Usuario, int, error) {
	return nil, 0, nil
}
func (r *usuarioHandlerRepositoryFake) SalvarAuth(context.Context, *usuario.UsuarioAuth) error {
	return nil
}
func (r *usuarioHandlerRepositoryFake) BuscarAuthPorProvider(context.Context, string, string) (*usuario.UsuarioAuth, error) {
	return nil, usuario.ErrNaoEncontrado
}
func (r *usuarioHandlerRepositoryFake) SalvarRefreshToken(context.Context, *usuario.RefreshToken) error {
	return nil
}
func (r *usuarioHandlerRepositoryFake) BuscarRefreshTokenPorHash(context.Context, string) (*usuario.RefreshToken, error) {
	return nil, usuario.ErrNaoAutorizado
}
func (r *usuarioHandlerRepositoryFake) RevogarRefreshToken(context.Context, string, time.Time) error {
	return nil
}
func (r *usuarioHandlerRepositoryFake) RevogarRefreshTokensDoUsuario(context.Context, string, time.Time) error {
	return nil
}
func (r *usuarioHandlerRepositoryFake) ExpurgarRefreshTokens(context.Context, time.Time) error {
	return nil
}
func (r *usuarioHandlerRepositoryFake) SalvarResetSenhaToken(context.Context, *usuario.ResetSenhaToken) error {
	return nil
}
func (r *usuarioHandlerRepositoryFake) BuscarResetSenhaTokenPorHash(context.Context, string) (*usuario.ResetSenhaToken, error) {
	return nil, usuario.ErrNaoEncontrado
}
func (r *usuarioHandlerRepositoryFake) MarcarResetSenhaTokenUsado(context.Context, string, time.Time) error {
	return nil
}

type clienteHandlerRepositoryFake struct {
	cliente    *cliente.Cliente
	atualizado *cliente.Cliente
}

func (r *clienteHandlerRepositoryFake) Salvar(context.Context, *cliente.Cliente) error { return nil }
func (r *clienteHandlerRepositoryFake) SalvarComVinculo(context.Context, *cliente.Cliente, *cliente.UsuarioCliente) error {
	return nil
}
func (r *clienteHandlerRepositoryFake) BuscarPorID(context.Context, string) (*cliente.Cliente, error) {
	if r.cliente == nil {
		return nil, cliente.ErrNaoEncontrado
	}
	return r.cliente, nil
}
func (r *clienteHandlerRepositoryFake) Atualizar(_ context.Context, c *cliente.Cliente) error {
	r.atualizado = c
	return nil
}
func (r *clienteHandlerRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *clienteHandlerRepositoryFake) Listar(context.Context, cliente.FiltroListagem) ([]*cliente.Cliente, int, error) {
	return nil, 0, nil
}
func (r *clienteHandlerRepositoryFake) VincularUsuario(context.Context, *cliente.UsuarioCliente) error {
	return nil
}
func (r *clienteHandlerRepositoryFake) ListarPorUsuario(context.Context, string) ([]*cliente.Cliente, error) {
	return nil, nil
}
func (r *clienteHandlerRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return nil
}

type leadHandlerRepositoryFake struct {
	salvo *lead.Lead
}

func (r *leadHandlerRepositoryFake) Salvar(_ context.Context, l *lead.Lead) error {
	l.Id = "lead-http"
	r.salvo = l
	return nil
}
func (r *leadHandlerRepositoryFake) BuscarPorID(context.Context, string) (*lead.Lead, error) {
	return nil, lead.ErrNaoEncontrado
}
func (r *leadHandlerRepositoryFake) BuscarDuplicado(context.Context, string, string, time.Time) (*lead.Lead, error) {
	return nil, lead.ErrNaoEncontrado
}
func (r *leadHandlerRepositoryFake) AtualizarStatus(context.Context, string, lead.AtualizacaoStatus, time.Time) error {
	return nil
}
func (r *leadHandlerRepositoryFake) Listar(context.Context, lead.FiltroListagem) ([]*lead.Lead, int, error) {
	return nil, 0, nil
}
func (r *leadHandlerRepositoryFake) Anonimizar(context.Context, string, time.Time) error { return nil }
func (r *leadHandlerRepositoryFake) AnonimizarExpirados(context.Context, time.Time) (int, error) {
	return 0, nil
}

type leadHandlerLandingPageFake struct {
	lp *landingpage.LandingPage
}

func (f *leadHandlerLandingPageFake) Salvar(context.Context, *landingpage.LandingPage) error {
	return nil
}
func (f *leadHandlerLandingPageFake) BuscarPorID(context.Context, string) (*landingpage.LandingPage, error) {
	if f.lp == nil {
		return nil, landingpage.ErrNaoEncontrada
	}
	return f.lp, nil
}
func (f *leadHandlerLandingPageFake) Atualizar(context.Context, *landingpage.LandingPage) error {
	return nil
}
func (f *leadHandlerLandingPageFake) SoftDelete(context.Context, string) error { return nil }
func (f *leadHandlerLandingPageFake) Listar(context.Context, landingpage.FiltroListagem) ([]*landingpage.LandingPage, int, error) {
	return nil, 0, nil
}
func (f *leadHandlerLandingPageFake) ContarPorCliente(context.Context, string) (int, error) {
	return 0, nil
}
func (f *leadHandlerLandingPageFake) AtualizarAutosave(context.Context, string, []byte, int) (int, error) {
	return 0, nil
}
func (f *leadHandlerLandingPageFake) AtualizarLock(context.Context, string, *string) error {
	return nil
}
func (f *leadHandlerLandingPageFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return nil
}
func (f *leadHandlerLandingPageFake) AtualizarPublicacao(context.Context, string, string, string, time.Time) error {
	return nil
}
func (f *leadHandlerLandingPageFake) AtualizarDespublicacao(context.Context, string) error {
	return nil
}

type publisherHandlerFake struct{}

func (publisherHandlerFake) Publicar(context.Context, evento.Evento) error { return nil }

type dashboardHandlerRepositoryFake struct {
	filtro dashboard.FiltroPeriodo
}

func (r *dashboardHandlerRepositoryFake) ObterResumoCliente(_ context.Context, _ string, filtro dashboard.FiltroPeriodo) (*dashboard.ResumoCliente, error) {
	r.filtro = filtro
	return &dashboard.ResumoCliente{ClienteId: "cliente-1", LandingPagesCriadas: 2, LandingPageMaisUsada: &dashboard.LandingPageUtilizada{Id: "lp-1", Nome: "LP", TotalExportacoes: 4}}, nil
}
func (r *dashboardHandlerRepositoryFake) ObterResumoUsuario(_ context.Context, _ string, filtro dashboard.FiltroPeriodo) (*dashboard.ResumoUsuario, error) {
	r.filtro = filtro
	return &dashboard.ResumoUsuario{UsuarioId: "usuario-1", ClientesVinculados: 1, LandingPagesCriadas: 2, LandingPageMaisUsada: &dashboard.LandingPageUtilizada{Id: "lp-1", Nome: "LP", TotalExportacoes: 4}}, nil
}
func (r *dashboardHandlerRepositoryFake) ObterResumoAdmin(_ context.Context, filtro dashboard.FiltroPeriodo) (*dashboard.ResumoAdmin, error) {
	r.filtro = filtro
	return &dashboard.ResumoAdmin{ClientesAtivos: 1}, nil
}

type clienteAcessoHandlerRepositoryFake struct{ err error }

func (r *clienteAcessoHandlerRepositoryFake) Salvar(context.Context, *cliente.Cliente) error {
	return nil
}
func (r *clienteAcessoHandlerRepositoryFake) SalvarComVinculo(context.Context, *cliente.Cliente, *cliente.UsuarioCliente) error {
	return nil
}
func (r *clienteAcessoHandlerRepositoryFake) BuscarPorID(context.Context, string) (*cliente.Cliente, error) {
	return nil, cliente.ErrNaoEncontrado
}
func (r *clienteAcessoHandlerRepositoryFake) Atualizar(context.Context, *cliente.Cliente) error {
	return nil
}
func (r *clienteAcessoHandlerRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *clienteAcessoHandlerRepositoryFake) Listar(context.Context, cliente.FiltroListagem) ([]*cliente.Cliente, int, error) {
	return nil, 0, nil
}
func (r *clienteAcessoHandlerRepositoryFake) VincularUsuario(context.Context, *cliente.UsuarioCliente) error {
	return nil
}
func (r *clienteAcessoHandlerRepositoryFake) ListarPorUsuario(context.Context, string) ([]*cliente.Cliente, error) {
	return nil, nil
}
func (r *clienteAcessoHandlerRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return r.err
}

func TestUsuarioHandlerIntegrationCriarUsuarioJSON(t *testing.T) {
	repository := &usuarioHandlerRepositoryFake{}
	handler := NewUsuarioHandler(ucusuario.NewUseCase(repository))
	request := httptest.NewRequest(http.MethodPost, "/usuarios", strings.NewReader(`{"nome":"Usuario HTTP","email":"http@sognix.test","senha":"senha-segura"}`))
	response := httptest.NewRecorder()

	handler.CriarUsuario(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status esperado %d, recebido %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var output dto.UsuarioOutput
	if err := json.NewDecoder(response.Body).Decode(&output); err != nil {
		t.Fatalf("decodificar resposta: %v", err)
	}
	if output.Id != "usuario-http" || output.Email != "http@sognix.test" || repository.salvo == nil {
		t.Fatalf("usuario criado inesperado: output=%+v salvo=%+v", output, repository.salvo)
	}
}

func TestClienteHandlerIntegrationAtualizarConflitoRetorna409(t *testing.T) {
	repository := &clienteHandlerRepositoryFake{cliente: &cliente.Cliente{Id: "cliente-1", Nome: "Cliente", Status: cliente.StatusAtivo, Version: 2}}
	handler := NewClienteHandler(uccliente.NewUseCase(repository))
	router := chi.NewRouter()
	router.Put("/clientes/{id}", handler.AtualizarCliente)

	request := httptest.NewRequest(http.MethodPut, "/clientes/cliente-1", strings.NewReader(`{"nome":"Novo","version":1}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status esperado %d, recebido %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if repository.atualizado != nil {
		t.Fatal("cliente com conflito nao deveria ser atualizado pelo handler")
	}
}

func TestLeadHandlerIntegrationCapturaLeadComIPValidado(t *testing.T) {
	repository := &leadHandlerRepositoryFake{}
	lpFake := &leadHandlerLandingPageFake{lp: &landingpage.LandingPage{Id: "lp-1", ClienteId: "cliente-1", Status: landingpage.StatusPublicada}}
	handler := NewLeadHandler(uclead.NewUseCase(repository, lpFake, publisherHandlerFake{}))
	t.Setenv("TRUSTED_PROXY_COUNT", "1")
	request := httptest.NewRequest(http.MethodPost, "/leads", strings.NewReader(`{"landingpage_id":"lp-1","nome":"Lead HTTP","email":"lead@sognix.test","whatsapp":"11999999999","consentimento":true}`))
	request.RemoteAddr = "10.0.0.1:443"
	request.Header.Set("X-Forwarded-For", "203.0.113.10")
	request.Header.Set("User-Agent", "handler-test")
	response := httptest.NewRecorder()

	handler.CapturarLead(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status esperado %d, recebido %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if repository.salvo == nil || repository.salvo.IP != "203.0.113.10" || repository.salvo.UserAgent != "handler-test" {
		t.Fatalf("lead salvo inesperado: %+v", repository.salvo)
	}
}

func TestDashboardHandlerIntegrationAutenticado(t *testing.T) {
	jwtServico := infrajwt.NewServico("access-secret-handler", "refresh-secret-handler")
	token, _, err := jwtServico.GerarTokens("usuario-1", "usuario@sognix.test", "usuario")
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}
	repository := &dashboardHandlerRepositoryFake{}
	handler := NewDashboardHandler(ucdashboard.NewUseCase(repository, &clienteAcessoHandlerRepositoryFake{}))
	router := chi.NewRouter()
	router.With(middleware.Autenticar(jwtServico)).Get("/dashboard", handler.ObterDashboard)

	request := httptest.NewRequest(http.MethodGet, "/dashboard?cliente_id=cliente-1&landingpage_id=lp-1&data_inicio=2026-05-01&data_fim=2026-05-02", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status esperado %d, recebido %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"landingpages_criadas":2`) {
		t.Fatalf("dashboard nao retornou resumo esperado: %s", response.Body.String())
	}
	if repository.filtro.ClienteId != "cliente-1" || repository.filtro.LandingPageId != "lp-1" || repository.filtro.DataFim != "2026-05-02" {
		t.Fatalf("filtros de dashboard nao propagados: %+v", repository.filtro)
	}
}

func TestDashboardHandlerIntegrationDataInvalidaRetorna400(t *testing.T) {
	jwtServico := infrajwt.NewServico("access-secret-handler", "refresh-secret-handler")
	token, _, err := jwtServico.GerarTokens("usuario-1", "usuario@sognix.test", "usuario")
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}
	handler := NewDashboardHandler(ucdashboard.NewUseCase(&dashboardHandlerRepositoryFake{}, &clienteAcessoHandlerRepositoryFake{}))
	router := chi.NewRouter()
	router.With(middleware.Autenticar(jwtServico)).Get("/dashboard", handler.ObterDashboard)

	request := httptest.NewRequest(http.MethodGet, "/dashboard?data_fim=02-05-2026", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status esperado %d, recebido %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}
