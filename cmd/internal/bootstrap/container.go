// Package bootstrap centraliza a inicializacao de dependencias compartilhadas pelos entrypoints.
package bootstrap

import (
	"context"
	"fmt"
	"io"

	ucauth "buildnix-backend/internal/app/usecase/auth"
	uccategoria "buildnix-backend/internal/app/usecase/categoria"
	uccliente "buildnix-backend/internal/app/usecase/cliente"
	ucdashboard "buildnix-backend/internal/app/usecase/dashboard"
	ucexportacao "buildnix-backend/internal/app/usecase/exportacao"
	uclandingpage "buildnix-backend/internal/app/usecase/landingpage"
	uclead "buildnix-backend/internal/app/usecase/lead"
	ucnegocio "buildnix-backend/internal/app/usecase/negocio"
	ucprospecto "buildnix-backend/internal/app/usecase/prospecto"
	uctemplate "buildnix-backend/internal/app/usecase/template"
	ucusuario "buildnix-backend/internal/app/usecase/usuario"
	"buildnix-backend/internal/domain/auditoria"
	"buildnix-backend/internal/domain/categoria"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/dashboard"
	"buildnix-backend/internal/domain/evento"
	"buildnix-backend/internal/domain/exportacao"
	"buildnix-backend/internal/domain/landingpage"
	"buildnix-backend/internal/domain/lead"
	"buildnix-backend/internal/domain/negocio"
	"buildnix-backend/internal/domain/plano"
	"buildnix-backend/internal/domain/prospecto"
	"buildnix-backend/internal/domain/template"
	"buildnix-backend/internal/domain/usuario"
	"buildnix-backend/internal/infra/config"
	"buildnix-backend/internal/infra/database/postgres"
	infrajwt "buildnix-backend/internal/infra/jwt"
)

// Container centraliza todas as dependencias da aplicacao.
type Container struct {
	// Infra
	Config     *config.Config
	Database   *postgres.Postgres
	Publisher  evento.Publisher
	JWTServico *infrajwt.Servico
	LogCloser  io.Closer

	// Repositories
	UsuarioRepository     usuario.Repository
	ClienteRepository     cliente.Repository
	DashboardRepository   dashboard.Repository
	NegocioRepository     negocio.Repository
	CategoriaRepository   categoria.Repository
	TemplateRepository    template.Repository
	LandingPageRepository landingpage.Repository
	LeadRepository        lead.Repository
	ProspectoRepository   prospecto.Repository
	ExportacaoRepository  exportacao.Repository
	AuditoriaRepository   auditoria.Repository
	EventOutboxRepository evento.OutboxRepository
	PlanoRepository       plano.Repository

	// App/Use cases
	AuthUseCase        *ucauth.UseCase
	UsuarioUseCase     *ucusuario.UseCase
	ClienteUseCase     *uccliente.UseCase
	DashboardUseCase   *ucdashboard.UseCase
	NegocioUseCase     *ucnegocio.UseCase
	CategoriaUseCase   *uccategoria.UseCase
	TemplateUseCase    *uctemplate.UseCase
	LandingPageUseCase *uclandingpage.UseCase
	LeadUseCase        *uclead.UseCase
	ProspectoUseCase   *ucprospecto.UseCase
	ExportacaoUseCase  *ucexportacao.UseCase
}

// New inicializa as dependencias na ordem obrigatoria: infra, repositorios e app.
func New(ctx context.Context) (*Container, error) {
	container := &Container{}

	if err := container.initInfra(ctx); err != nil {
		return nil, fmt.Errorf("inicializacao de infra: %w", err)
	}
	if err := container.initRepositories(); err != nil {
		container.Close()
		return nil, fmt.Errorf("inicializacao de repositorios: %w", err)
	}
	if err := container.seedAdmin(ctx); err != nil {
		container.Close()
		return nil, fmt.Errorf("seed admin: %w", err)
	}
	if err := container.initApp(); err != nil {
		container.Close()
		return nil, fmt.Errorf("inicializacao da app: %w", err)
	}

	return container, nil
}

// Close libera recursos alocados pelo container.
func (c *Container) Close() {
	if c.Database != nil {
		c.Database.Close()
	}
	if c.LogCloser != nil {
		_ = c.LogCloser.Close()
	}
}
