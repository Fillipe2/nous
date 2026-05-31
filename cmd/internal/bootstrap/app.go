package bootstrap

import (
	ucauth "nous/internal/app/usecase/auth"
	uccategoria "nous/internal/app/usecase/categoria"
	uccliente "nous/internal/app/usecase/cliente"
	ucdashboard "nous/internal/app/usecase/dashboard"
	ucexportacao "nous/internal/app/usecase/exportacao"
	uclandingpage "nous/internal/app/usecase/landingpage"
	uclead "nous/internal/app/usecase/lead"
	ucnegocio "nous/internal/app/usecase/negocio"
	ucfaturamento "nous/internal/app/usecase/faturamento"
	ucleadaquisicao "nous/internal/app/usecase/lead_aquisicao"
	uctemplate "nous/internal/app/usecase/template"
	ucusuario "nous/internal/app/usecase/usuario"
	"nous/internal/infra/eventbus"
	"nous/internal/infra/hosting/noop"
)

func (c *Container) initApp() error {
	c.Publisher = eventbus.NewOutboxPublisher(c.EventOutboxRepository, c.Publisher)
	c.Publisher = eventbus.NewFaturamentoPublisher(c.Publisher, c.UsoRepository, c.BiRepository)

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
	c.LeadAquisicaoUseCase = ucleadaquisicao.NewUseCase(c.LeadAquisicaoRepository, c.Publisher)
	c.FaturamentoUseCase = ucfaturamento.NewUseCase(c.UsoRepository, c.FaturamentoRepository, c.BiRepository)
	c.ExportacaoUseCase = ucexportacao.NewUseCase(c.ExportacaoRepository, c.LandingPageRepository)

	return nil
}
