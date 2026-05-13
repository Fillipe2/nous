package bootstrap

import (
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
	"buildnix-backend/internal/infra/eventbus"
	"buildnix-backend/internal/infra/hosting/noop"
)

func (c *Container) initApp() error {
	c.Publisher = eventbus.NewOutboxPublisher(c.EventOutboxRepository, c.Publisher)

	if c.Deployer == nil {
		c.Deployer = noop.New()
	}

	c.AuthUseCase = ucauth.NewUseCase(c.UsuarioRepository, c.ClienteRepository, c.JWTServico)
	c.UsuarioUseCase = ucusuario.NewUseCase(c.UsuarioRepository)
	c.ClienteUseCase = uccliente.NewUseCase(c.ClienteRepository)
	c.DashboardUseCase = ucdashboard.NewUseCase(c.DashboardRepository, c.ClienteRepository)
	c.NegocioUseCase = ucnegocio.NewUseCase(c.NegocioRepository, c.ClienteRepository)
	c.CategoriaUseCase = uccategoria.NewUseCase(c.CategoriaRepository)
	c.TemplateUseCase = uctemplate.NewUseCase(c.TemplateRepository, c.ClienteRepository, c.PlanoRepository)
	c.LandingPageUseCase = uclandingpage.NewUseCase(c.LandingPageRepository, c.ClienteRepository, c.Publisher, c.PlanoRepository, c.TemplateRepository, c.Deployer, c.ExportacaoRepository)
	c.LeadUseCase = uclead.NewUseCase(c.LeadRepository, c.LandingPageRepository, c.Publisher)
	c.ProspectoUseCase = ucprospecto.NewUseCase(c.ProspectoRepository)
	c.ExportacaoUseCase = ucexportacao.NewUseCase(c.ExportacaoRepository, c.LandingPageRepository)

	return nil
}
