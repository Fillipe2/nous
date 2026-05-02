package router

import (
	_ "embed"
	"net/http"

	"buildnix-backend/internal/domain/auditoria"
	infrajwt "buildnix-backend/internal/infra/jwt"
	"buildnix-backend/internal/interface/webapi/internal/handler"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

//go:embed openapi.yaml
var openapiYAML []byte

type Handlers struct {
	Auth        *handler.AuthHandler
	Usuario     *handler.UsuarioHandler
	Cliente     *handler.ClienteHandler
	Dashboard   *handler.DashboardHandler
	Negocio     *handler.NegocioHandler
	Categoria   *handler.CategoriaHandler
	Template    *handler.TemplateHandler
	LandingPage *handler.LandingPageHandler
	Lead        *handler.LeadHandler
	Exportacao  *handler.ExportacaoHandler
	Auditoria   *handler.AuditoriaHandler
	Opcao       *handler.OpcaoHandler
}

func NovoRouter(handlers *Handlers, jwtServico *infrajwt.Servico, auditoriaRepo auditoria.Repository) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.Identify)
	r.Use(middleware.HeadersSeguranca)
	r.Use(middleware.LimitarTaxa)
	r.Use(middleware.LimitarBody)
	r.Use(cors())
	r.Get("/swagger/openapi.yaml", servirOpenAPI)

	autenticar := middleware.Autenticar(jwtServico)

	// Rotas públicas
	r.Group(func(r chi.Router) {
		r.Post("/auth/login", handlers.Auth.Login)
		r.Post("/auth/refresh", handlers.Auth.RefreshToken)
		r.Post("/auth/logout", handlers.Auth.Logout)
		r.Get("/auth/google/login", handlers.Auth.GoogleLogin)
		r.Get("/auth/google/callback", handlers.Auth.GoogleCallback)

		r.Post("/usuarios", handlers.Usuario.CriarUsuario)

		// Lead público (formulário da LP)
		r.Post("/leads", handlers.Lead.CapturarLead)
	})

	// Rotas autenticadas
	r.Group(func(r chi.Router) {
		r.Use(autenticar)
		r.Use(middleware.RegistrarAuditoria(auditoriaRepo))

		// Usuários
		r.Get("/usuarios/me", handlers.Usuario.ObterUsuarioAutenticado)
		r.Put("/usuarios/{id}", handlers.Usuario.AtualizarUsuario)
		r.Delete("/usuarios/{id}", handlers.Usuario.ExcluirUsuario)
		r.With(middleware.ApenasAdmin).Get("/usuarios", handlers.Usuario.ListarUsuarios)

		// Clientes
		r.Get("/clientes/me", handlers.Cliente.ListarClientesDoUsuario)
		r.With(middleware.ApenasAdmin).Get("/clientes", handlers.Cliente.ListarClientes)
		r.With(middleware.ApenasAdmin).Post("/clientes", handlers.Cliente.CriarCliente)
		r.With(middleware.ApenasAdmin).Put("/clientes/{id}", handlers.Cliente.AtualizarCliente)
		r.With(middleware.ApenasAdmin).Delete("/clientes/{id}", handlers.Cliente.ExcluirCliente)
		r.With(middleware.ApenasAdmin).Post("/usuarios-clientes", handlers.Cliente.VincularUsuario)

		// Dashboard
		r.Get("/dashboard", handlers.Dashboard.ObterDashboardCliente)
		r.With(middleware.ApenasAdmin).Get("/dashboard/admin", handlers.Dashboard.ObterDashboardAdmin)

		// Opções para combos/filtros do frontend
		r.Get("/opcoes/clientes", handlers.Opcao.ListarClientes)
		r.With(middleware.ApenasAdmin).Get("/opcoes/usuarios", handlers.Opcao.ListarUsuarios)
		r.Get("/opcoes/negocios", handlers.Opcao.ListarNegocios)
		r.Get("/opcoes/categorias", handlers.Opcao.ListarCategorias)
		r.Get("/opcoes/templates", handlers.Opcao.ListarTemplates)
		r.Get("/opcoes/landingpages", handlers.Opcao.ListarLandingPages)
		r.Get("/opcoes/enums", handlers.Opcao.ListarEnums)

		// Negócios
		r.Get("/negocios", handlers.Negocio.ListarNegocios)
		r.Post("/negocios", handlers.Negocio.CriarNegocio)
		r.Put("/negocios/{id}", handlers.Negocio.AtualizarNegocio)
		r.Delete("/negocios/{id}", handlers.Negocio.ExcluirNegocio)

		// Categorias
		r.Get("/categorias", handlers.Categoria.ListarCategorias)
		r.With(middleware.ApenasAdmin).Post("/categorias", handlers.Categoria.CriarCategoria)
		r.With(middleware.ApenasAdmin).Put("/categorias/{id}", handlers.Categoria.AtualizarCategoria)
		r.With(middleware.ApenasAdmin).Delete("/categorias/{id}", handlers.Categoria.ExcluirCategoria)

		// Templates
		r.Get("/templates", handlers.Template.ListarTemplates)
		r.Post("/templates", handlers.Template.UploadTemplate)
		r.With(middleware.ApenasAdmin).Delete("/templates/{id}", handlers.Template.ExcluirTemplate)

		// Landing Pages
		r.Get("/landingpages", handlers.LandingPage.ListarLandingPages)
		r.Post("/landingpages", handlers.LandingPage.CriarLandingPage)
		r.Get("/landingpages/{id}", handlers.LandingPage.ObterLandingPage)
		r.Put("/landingpages/{id}", handlers.LandingPage.AtualizarLandingPage)
		r.Delete("/landingpages/{id}", handlers.LandingPage.ExcluirLandingPage)
		r.Patch("/landingpages/{id}/autosave", handlers.LandingPage.Autosave)
		r.Patch("/landingpages/{id}/lock", handlers.LandingPage.BloquearEdicao)
		r.Patch("/landingpages/{id}/unlock", handlers.LandingPage.DesbloquearEdicao)
		r.Post("/landingpages/{id}/exportar", handlers.Exportacao.ExportarLandingPage)

		// Leads: dados pessoais ficam restritos ao admin/Sognix
		r.With(middleware.ApenasAdmin).Get("/leads", handlers.Lead.ListarLeads)
		r.With(middleware.ApenasAdmin).Get("/leads/{id}", handlers.Lead.ObterLead)
		r.With(middleware.ApenasAdmin).Patch("/leads/{id}/status", handlers.Lead.AtualizarStatusLead)

		// Exportações
		r.Get("/exportacoes", handlers.Exportacao.ListarExportacoes)
		r.Get("/exportacoes/{id}/download", handlers.Exportacao.BaixarExportacao)

		// Auditoria
		r.With(middleware.ApenasAdmin).Get("/auditoria", handlers.Auditoria.ListarAuditoria)
		r.With(middleware.ApenasAdmin).Get("/auditoria/{id}", handlers.Auditoria.ObterAuditoria)
	})

	return r
}

func servirOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openapiYAML)
}

func cors() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
