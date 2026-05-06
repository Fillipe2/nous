package router

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"buildnix-backend/internal/domain/auditoria"
	infrajwt "buildnix-backend/internal/infra/jwt"
	"buildnix-backend/internal/interface/webapi/internal/handler"
	"buildnix-backend/internal/interface/webapi/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

//go:embed openapi.yaml
var openapiYAML []byte

const swaggerHTML = `<!doctype html>
<html lang="pt-BR">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Buildnix API</title>
	<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' rx='14' fill='%23111827'/%3E%3Cpath d='M18 20h12a8 8 0 0 1 0 16H18zm0 16h14a8 8 0 0 1 0 12H18zM38 20h8v28h-8z' fill='%2338bdf8'/%3E%3C/svg%3E">
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
	<style>
		:root {
			--bn-bg: #0f172a;
			--bn-bg-soft: #111827;
			--bn-surface: #ffffff;
			--bn-border: #e5e7eb;
			--bn-text: #0f172a;
			--bn-muted: #64748b;
			--bn-accent: #38bdf8;
			--bn-accent-strong: #0ea5e9;
		}
		* { box-sizing: border-box; }
		html, body { margin: 0; padding: 0; background: #f8fafc; color: var(--bn-text); font-family: 'Inter', system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif; }
		.bn-topbar {
			background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
			color: #f8fafc;
			padding: 18px 32px;
			display: flex;
			align-items: center;
			justify-content: space-between;
			border-bottom: 1px solid #1e293b;
			box-shadow: 0 1px 3px rgba(0,0,0,.18);
		}
		.bn-brand { display: flex; align-items: center; gap: 14px; }
		.bn-brand-mark {
			width: 40px; height: 40px; border-radius: 10px;
			background: linear-gradient(135deg, #38bdf8 0%, #6366f1 100%);
			display: grid; place-items: center;
			color: #0f172a; font-weight: 800; font-size: 18px; letter-spacing: -0.5px;
		}
		.bn-brand-text { display: flex; flex-direction: column; }
		.bn-brand-title { font-weight: 700; font-size: 16px; letter-spacing: 0.2px; }
		.bn-brand-sub { font-size: 12px; color: #94a3b8; }
		.bn-actions { display: flex; gap: 10px; align-items: center; }
		.bn-link {
			color: #cbd5e1; text-decoration: none; font-size: 13px;
			padding: 6px 12px; border-radius: 8px; border: 1px solid #334155;
			transition: all .15s ease;
		}
		.bn-link:hover { background: #1e293b; color: #f8fafc; border-color: #475569; }

		#swagger-ui { max-width: 1200px; margin: 0 auto; padding: 24px 16px 64px; }
		.swagger-ui .topbar { display: none; }
		.swagger-ui .info { margin: 24px 0 32px; }
		.swagger-ui .info .title { font-size: 28px; font-weight: 700; color: var(--bn-text); }
		.swagger-ui .info .title small { display: none; }
		.swagger-ui .info .title small.version-stamp {
			display: inline-block; background: #0ea5e9; color: #fff; padding: 2px 10px;
			border-radius: 999px; font-size: 11px; margin-left: 10px; vertical-align: middle;
		}
		.swagger-ui .scheme-container {
			background: #fff; border: 1px solid var(--bn-border);
			border-radius: 12px; padding: 16px 20px; box-shadow: 0 1px 2px rgba(15,23,42,.04);
			margin: 0 0 24px;
		}
		.swagger-ui .opblock-tag {
			border-bottom: 1px solid var(--bn-border);
			padding: 14px 12px; font-size: 18px; font-weight: 600;
		}
		.swagger-ui .opblock-tag:hover { background: #f1f5f9; }
		.swagger-ui .opblock-tag small { color: var(--bn-muted); font-weight: 400; }
		.swagger-ui .opblock {
			border-radius: 10px; border: 1px solid var(--bn-border);
			margin: 0 0 10px; box-shadow: none;
		}
		.swagger-ui .opblock .opblock-summary { padding: 8px 16px; }
		.swagger-ui .opblock .opblock-summary-method {
			border-radius: 6px; min-width: 72px; font-weight: 700; font-size: 12px;
		}
		.swagger-ui .opblock .opblock-summary-path { font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 14px; }
		.swagger-ui .opblock .opblock-summary-description { color: var(--bn-muted); font-size: 13px; }
		.swagger-ui .opblock-get  { background: #f0f9ff; border-color: #bae6fd; }
		.swagger-ui .opblock-post { background: #f0fdf4; border-color: #bbf7d0; }
		.swagger-ui .opblock-put  { background: #fff7ed; border-color: #fed7aa; }
		.swagger-ui .opblock-patch { background: #fefce8; border-color: #fde68a; }
		.swagger-ui .opblock-delete { background: #fef2f2; border-color: #fecaca; }
		.swagger-ui .btn.authorize {
			background: var(--bn-accent-strong); color: #fff; border-color: var(--bn-accent-strong);
			border-radius: 8px;
		}
		.swagger-ui .btn.authorize:hover { background: #0284c7; }
		.swagger-ui .btn.authorize svg { fill: #fff; }
		.swagger-ui .btn.execute {
			background: var(--bn-text); border-color: var(--bn-text);
			border-radius: 8px;
		}
		.swagger-ui .filter .operation-filter-input {
			border-radius: 10px; border: 1px solid var(--bn-border);
			padding: 10px 14px; font-size: 14px;
		}
		.swagger-ui .filter .operation-filter-input:focus {
			border-color: var(--bn-accent-strong);
			box-shadow: 0 0 0 3px rgba(14,165,233,.18);
			outline: none;
		}
		.swagger-ui .model-box, .swagger-ui section.models {
			border-radius: 10px; border: 1px solid var(--bn-border);
		}
		.swagger-ui section.models.is-open h4 { border-bottom: 1px solid var(--bn-border); }
		.swagger-ui table tbody tr td { padding: 10px; }
		.swagger-ui .response-col_status { font-weight: 600; }
		.swagger-ui .markdown code, .swagger-ui .renderedMarkdown code {
			background: #f1f5f9; color: #0f172a; border-radius: 4px; padding: 1px 6px;
		}
	</style>
</head>
<body>
	<header class="bn-topbar">
		<div class="bn-brand">
			<div class="bn-brand-mark">B</div>
			<div class="bn-brand-text">
				<span class="bn-brand-title">Buildnix Backend</span>
				<span class="bn-brand-sub">Documentação interativa da API</span>
			</div>
		</div>
		<nav class="bn-actions">
			<a class="bn-link" href="/swagger/openapi.yaml" target="_blank" rel="noopener">openapi.yaml</a>
			<a class="bn-link" href="/health" target="_blank" rel="noopener">health</a>
		</nav>
	</header>
	<main id="swagger-ui"></main>
	<script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
	<script>
		window.onload = function () {
			window.ui = SwaggerUIBundle({
				url: "/swagger/openapi.yaml",
				dom_id: "#swagger-ui",
				deepLinking: true,
				persistAuthorization: true,
				displayRequestDuration: true,
				tryItOutEnabled: true,
				filter: true,
				docExpansion: "none",
				defaultModelsExpandDepth: -1,
				defaultModelExpandDepth: 2,
				operationsSorter: "alpha",
				tagsSorter: "alpha",
				syntaxHighlight: { activate: true, theme: "agate" }
			});
		};
	</script>
</body>
</html>`

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
	Prospecto   *handler.ProspectoHandler
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
	r.Get("/health", servirHealth)
	r.Get("/swagger", servirSwaggerUI)
	r.Get("/swagger/", servirSwaggerUI)
	r.Get("/swagger/openapi.yaml", servirOpenAPI)
	r.Get("/.well-known/jwks.json", servirJWKS(jwtServico))

	autenticar := middleware.Autenticar(jwtServico)

	// Rotas públicas
	r.Group(func(r chi.Router) {
		r.Post("/auth/login", handlers.Auth.Login)
		r.Post("/auth/refresh", handlers.Auth.RefreshToken)
		r.Post("/auth/logout", handlers.Auth.Logout)
		r.Post("/auth/esqueci-senha", handlers.Auth.EsqueciSenha)
		r.Post("/auth/redefinir-senha", handlers.Auth.RedefinirSenha)
		r.Get("/auth/google/login", handlers.Auth.GoogleLogin)
		r.Get("/auth/google/callback", handlers.Auth.GoogleCallback)

		r.Post("/usuarios", handlers.Usuario.CriarUsuario)

		// Lead público (formulário da LP)
		r.Post("/leads", handlers.Lead.CapturarLead)

		// Prospecto público (formulário de aquisição Sognix B2B)
		r.Post("/prospectos", handlers.Prospecto.RegistrarProspecto)
	})

	// Rotas autenticadas
	r.Group(func(r chi.Router) {
		r.Use(autenticar)
		r.Use(middleware.RegistrarAuditoria(auditoriaRepo))

		// Usuários
		r.Get("/usuarios/me", handlers.Usuario.ObterUsuarioAutenticado)
		r.Patch("/usuarios/me/senha", handlers.Usuario.AlterarSenha)
		r.Put("/usuarios/{id}", handlers.Usuario.AtualizarUsuario)
		r.Delete("/usuarios/{id}", handlers.Usuario.ExcluirUsuario)
		r.With(middleware.ApenasAdmin).Get("/usuarios", handlers.Usuario.ListarUsuarios)

		// Clientes
		r.Get("/clientes/me", handlers.Cliente.ListarClientesDoUsuario)
		r.With(middleware.ApenasAdmin).Get("/clientes", handlers.Cliente.ListarClientes)
		r.Post("/clientes", handlers.Cliente.CriarCliente)
		r.With(middleware.ApenasAdmin).Put("/clientes/{id}", handlers.Cliente.AtualizarCliente)
		r.With(middleware.ApenasAdmin).Delete("/clientes/{id}", handlers.Cliente.ExcluirCliente)

		// Dashboard
		r.Get("/dashboard", handlers.Dashboard.ObterDashboard)

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
		r.With(middleware.ApenasAdmin).Post("/leads/{id}/anonimizar", handlers.Lead.AnonimizarLead)
		r.With(middleware.ApenasAdmin).Post("/leads/anonimizar-expirados", handlers.Lead.AnonimizarLeadsExpirados)

		// Exportações
		r.Get("/exportacoes", handlers.Exportacao.ListarExportacoes)
		r.Get("/exportacoes/{id}/download", handlers.Exportacao.BaixarExportacao)

		// Prospectos: dados B2B restritos ao admin/Sognix
		r.With(middleware.ApenasAdmin).Get("/prospectos", handlers.Prospecto.ListarProspectos)
		r.With(middleware.ApenasAdmin).Get("/prospectos/{id}", handlers.Prospecto.ObterProspecto)
		r.With(middleware.ApenasAdmin).Patch("/prospectos/{id}/funil", handlers.Prospecto.AtualizarFunilProspecto)

		// Auditoria
		r.With(middleware.ApenasAdmin).Get("/auditoria", handlers.Auditoria.ListarAuditoria)
		r.With(middleware.ApenasAdmin).Get("/auditoria/{id}", handlers.Auditoria.ObterAuditoria)
	})

	return r
}

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func servirJWKS(jwtServico *infrajwt.Servico) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		jwks, ok := jwtServico.JWKS()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if !ok {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"jwks nao configurado: JWT_RSA_PRIVATE_KEY ausente"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		data, err := jsonMarshal(jwks)
		if err != nil {
			log.Printf("servirJWKS encode: %v", err)
			return
		}
		_, _ = w.Write(data)
	}
}

func servirOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openapiYAML)
}

func servirSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerHTML))
}

func servirHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func cors() func(http.Handler) http.Handler {
	origensPermitidas := carregarOrigensCORS()
	log.Printf("cors origens permitidas: %s", strings.Join(origensPermitidas, ","))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && origemPermitida(origin, origensPermitidas) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "600")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func carregarOrigensCORS() []string {
	valor := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if valor == "" {
		return []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8080"}
	}
	origens := strings.Split(valor, ",")
	saida := make([]string, 0, len(origens))
	for _, origem := range origens {
		origem = strings.TrimSpace(origem)
		if origem != "" {
			saida = append(saida, origem)
		}
	}
	return saida
}

func origemPermitida(origin string, permitidas []string) bool {
	for _, permitida := range permitidas {
		if permitida == "*" || permitida == origin {
			return true
		}
	}
	return false
}
