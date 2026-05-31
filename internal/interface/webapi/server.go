// Package webapi implementa a API REST principal do Buildnix.
package webapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	ucauth "nous/internal/app/usecase/auth"
	uccategoria "nous/internal/app/usecase/categoria"
	uccliente "nous/internal/app/usecase/cliente"
	ucdashboard "nous/internal/app/usecase/dashboard"
	ucexportacao "nous/internal/app/usecase/exportacao"
	uclp "nous/internal/app/usecase/landingpage"
	uclead "nous/internal/app/usecase/lead"
	ucnegocio "nous/internal/app/usecase/negocio"
	ucleadaquisicao "nous/internal/app/usecase/lead_aquisicao"
	uctemplate "nous/internal/app/usecase/template"
	ucusuario "nous/internal/app/usecase/usuario"
	"nous/internal/domain/auditoria"
	infrajwt "nous/internal/infra/jwt"
	"nous/internal/interface/webapi/internal/handler"
	"nous/internal/interface/webapi/internal/middleware"
	"nous/internal/interface/webapi/internal/router"
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
	LeadUseCase           *uclead.UseCase
	LeadAquisicaoUseCase  *ucleadaquisicao.UseCase
	ExportacaoUseCase     *ucexportacao.UseCase
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
		LeadAquisicao: handler.NewLeadAquisicaoHandler(s.LeadAquisicaoUseCase),
		Exportacao:  handler.NewExportacaoHandler(s.ExportacaoUseCase),
		Auditoria:   handler.NewAuditoriaHandler(s.AuditoriaRepo),
		Opcao:       handler.NewOpcaoHandler(s.ClienteUseCase, s.UsuarioUseCase, s.NegocioUseCase, s.CategoriaUseCase, s.TemplateUseCase, s.LandingPageUseCase),
	}
	return router.NovoRouter(handlers, s.JWTServico, s.AuditoriaRepo)
}

func CloseSharedResources() {
	middleware.CloseRateLimiter()
}
