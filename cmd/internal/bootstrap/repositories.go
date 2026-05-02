package bootstrap

import (
	"fmt"

	"buildnix-backend/internal/infra/database/postgres"
)

func (c *Container) initRepositories() error {
	usuarioRepository, err := postgres.NewUsuarioRepository(c.Database)
	if err != nil {
		return fmt.Errorf("usuario repository: %w", err)
	}
	clienteRepository, err := postgres.NewClienteRepository(c.Database)
	if err != nil {
		return fmt.Errorf("cliente repository: %w", err)
	}
	dashboardRepository, err := postgres.NewDashboardRepository(c.Database)
	if err != nil {
		return fmt.Errorf("dashboard repository: %w", err)
	}
	negocioRepository, err := postgres.NewNegocioRepository(c.Database)
	if err != nil {
		return fmt.Errorf("negocio repository: %w", err)
	}
	categoriaRepository, err := postgres.NewCategoriaRepository(c.Database)
	if err != nil {
		return fmt.Errorf("categoria repository: %w", err)
	}
	templateRepository, err := postgres.NewTemplateRepository(c.Database)
	if err != nil {
		return fmt.Errorf("template repository: %w", err)
	}
	landingPageRepository, err := postgres.NewLandingPageRepository(c.Database)
	if err != nil {
		return fmt.Errorf("landingpage repository: %w", err)
	}
	leadRepository, err := postgres.NewLeadRepository(c.Database)
	if err != nil {
		return fmt.Errorf("lead repository: %w", err)
	}
	exportacaoRepository, err := postgres.NewExportacaoRepository(c.Database)
	if err != nil {
		return fmt.Errorf("exportacao repository: %w", err)
	}
	auditoriaRepository, err := postgres.NewAuditoriaRepository(c.Database)
	if err != nil {
		return fmt.Errorf("auditoria repository: %w", err)
	}
	eventOutboxRepository, err := postgres.NewEventOutboxRepository(c.Database)
	if err != nil {
		return fmt.Errorf("event outbox repository: %w", err)
	}
	planoRepository, err := postgres.NewPlanoRepository(c.Database)
	if err != nil {
		return fmt.Errorf("plano repository: %w", err)
	}

	c.UsuarioRepository = usuarioRepository
	c.ClienteRepository = clienteRepository
	c.DashboardRepository = dashboardRepository
	c.NegocioRepository = negocioRepository
	c.CategoriaRepository = categoriaRepository
	c.TemplateRepository = templateRepository
	c.LandingPageRepository = landingPageRepository
	c.LeadRepository = leadRepository
	c.ExportacaoRepository = exportacaoRepository
	c.AuditoriaRepository = auditoriaRepository
	c.EventOutboxRepository = eventOutboxRepository
	c.PlanoRepository = planoRepository

	return nil
}
