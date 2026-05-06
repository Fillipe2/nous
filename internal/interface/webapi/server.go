// Package webapi implementa a API REST principal do Buildnix.
package webapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	ucauth "buildnix-backend/internal/app/usecase/auth"
	uccategoria "buildnix-backend/internal/app/usecase/categoria"
	uccliente "buildnix-backend/internal/app/usecase/cliente"
	ucdashboard "buildnix-backend/internal/app/usecase/dashboard"
	ucexportacao "buildnix-backend/internal/app/usecase/exportacao"
	uclp "buildnix-backend/internal/app/usecase/landingpage"
	uclead "buildnix-backend/internal/app/usecase/lead"
	ucnegocio "buildnix-backend/internal/app/usecase/negocio"
	ucprospecto "buildnix-backend/internal/app/usecase/prospecto"
	uctemplate "buildnix-backend/internal/app/usecase/template"
	ucusuario "buildnix-backend/internal/app/usecase/usuario"
	"buildnix-backend/internal/domain/auditoria"
	infrajwt "buildnix-backend/internal/infra/jwt"
	"buildnix-backend/internal/interface/webapi/internal/handler"
	"buildnix-backend/internal/interface/webapi/internal/middleware"
	"buildnix-backend/internal/interface/webapi/internal/router"
)

// Server agrega as dependencias necessarias para configurar rotas, middlewares e handlers da API REST.
type Server struct {
	AuthUseCase        *ucauth.UseCase
	UsuarioUseCase     *ucusuario.UseCase
	ClienteUseCase     *uccliente.UseCase
	DashboardUseCase   *ucdashboard.UseCase
	NegocioUseCase     *ucnegocio.UseCase
	CategoriaUseCase   *uccategoria.UseCase
	TemplateUseCase    *uctemplate.UseCase
	LandingPageUseCase *uclp.UseCase
	LeadUseCase        *uclead.UseCase
	ProspectoUseCase   *ucprospecto.UseCase
	ExportacaoUseCase  *ucexportacao.UseCase
	AuditoriaRepo      auditoria.Repository
	JWTServico         *infrajwt.Servico
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

// Start configura o router HTTP e inicia o listener em goroutine, retornando o servidor para graceful shutdown.
func (s *Server) Start(ctx context.Context, address string) (*http.Server, error) {
	if s.JWTServico == nil {
		return nil, errors.New("servico JWT da webapi nao configurado")
	}
	if s.AuditoriaRepo == nil {
		return nil, errors.New("repositorio de auditoria da webapi nao configurado")
	}
	middleware.ConfigurarRateLimiter(os.Getenv("REDIS_URL"))

	httpServer := &http.Server{
		Addr:              address,
		Handler:           s.setupRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("erro no servidor HTTP: %v", err)
		}
	}()

	return httpServer, nil
}

func (s *Server) setupRouter() http.Handler {
	handlers := &router.Handlers{
		Auth:        handler.NewAuthHandler(s.AuthUseCase, s.GoogleClientID, s.GoogleClientSecret, s.GoogleRedirectURL),
		Usuario:     handler.NewUsuarioHandler(s.UsuarioUseCase),
		Cliente:     handler.NewClienteHandler(s.ClienteUseCase),
		Dashboard:   handler.NewDashboardHandler(s.DashboardUseCase),
		Negocio:     handler.NewNegocioHandler(s.NegocioUseCase),
		Categoria:   handler.NewCategoriaHandler(s.CategoriaUseCase),
		Template:    handler.NewTemplateHandler(s.TemplateUseCase),
		LandingPage: handler.NewLandingPageHandler(s.LandingPageUseCase),
		Lead:        handler.NewLeadHandler(s.LeadUseCase),
		Prospecto:   handler.NewProspectoHandler(s.ProspectoUseCase),
		Exportacao:  handler.NewExportacaoHandler(s.ExportacaoUseCase),
		Auditoria:   handler.NewAuditoriaHandler(s.AuditoriaRepo),
		Opcao:       handler.NewOpcaoHandler(s.ClienteUseCase, s.UsuarioUseCase, s.NegocioUseCase, s.CategoriaUseCase, s.TemplateUseCase, s.LandingPageUseCase),
	}
	return router.NovoRouter(handlers, s.JWTServico, s.AuditoriaRepo)
}

func CloseSharedResources() {
	middleware.CloseRateLimiter()
}
