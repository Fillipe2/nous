package dashboard

import (
	"context"
	"errors"
	"testing"

	"nous/internal/app/dto"
	domain "nous/internal/domain/dashboard"
	"nous/internal/domain/negocio"
)

type dashboardRepositoryFake struct {
	resumoNegocio *domain.ResumoNegocio
	resumoUsuario *domain.ResumoUsuario
	resumoAdmin   *domain.ResumoAdmin
	filtroUsuario domain.FiltroPeriodo
}

func (r *dashboardRepositoryFake) ObterResumoNegocio(context.Context, string, domain.FiltroPeriodo) (*domain.ResumoNegocio, error) {
	return r.resumoNegocio, nil
}

func (r *dashboardRepositoryFake) ObterResumoUsuario(_ context.Context, _ string, filtro domain.FiltroPeriodo) (*domain.ResumoUsuario, error) {
	r.filtroUsuario = filtro
	return r.resumoUsuario, nil
}

func (r *dashboardRepositoryFake) ObterResumoAdmin(context.Context, domain.FiltroPeriodo) (*domain.ResumoAdmin, error) {
	return r.resumoAdmin, nil
}

type negocioRepositoryFake struct{ err error }

func (r *negocioRepositoryFake) Salvar(context.Context, *negocio.Negocio) error { return nil }
func (r *negocioRepositoryFake) SalvarComVinculo(context.Context, *negocio.Negocio, *negocio.UsuarioNegocio) error {
	return nil
}
func (r *negocioRepositoryFake) BuscarPorID(context.Context, string) (*negocio.Negocio, error) {
	return nil, negocio.ErrNaoEncontrado
}
func (r *negocioRepositoryFake) Atualizar(context.Context, *negocio.Negocio) error { return nil }
func (r *negocioRepositoryFake) SoftDelete(context.Context, string) error          { return nil }
func (r *negocioRepositoryFake) Listar(context.Context, negocio.FiltroListagem) ([]*negocio.Negocio, int, error) {
	return nil, 0, nil
}
func (r *negocioRepositoryFake) VincularUsuario(context.Context, *negocio.UsuarioNegocio) error {
	return nil
}
func (r *negocioRepositoryFake) ListarPorUsuario(context.Context, string) ([]*negocio.Negocio, error) {
	return nil, nil
}
func (r *negocioRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return r.err
}

func TestObterDashboardNegocioExigeNegocioId(t *testing.T) {
	useCase := NewUseCase(&dashboardRepositoryFake{}, &negocioRepositoryFake{})

	_, err := useCase.ObterDashboardNegocio(context.Background(), "", domain.FiltroPeriodo{}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrNegocioObrigatorio) {
		t.Fatalf("erro esperado de negocio obrigatorio, recebido %v", err)
	}
}

func TestObterDashboardNegocioValidaAcessoUsuario(t *testing.T) {
	useCase := NewUseCase(&dashboardRepositoryFake{resumoNegocio: &domain.ResumoNegocio{NegocioId: "negocio-1"}}, &negocioRepositoryFake{err: negocio.ErrNaoAutorizado})

	_, err := useCase.ObterDashboardNegocio(context.Background(), "negocio-1", domain.FiltroPeriodo{}, "usuario-1", "usuario")

	if err == nil {
		t.Fatal("erro esperado de acesso ao negocio")
	}
}

func TestObterDashboardAdminExigeRoleAdmin(t *testing.T) {
	useCase := NewUseCase(&dashboardRepositoryFake{resumoAdmin: &domain.ResumoAdmin{}}, &negocioRepositoryFake{})

	_, err := useCase.ObterDashboardAdmin(context.Background(), domain.FiltroPeriodo{}, "usuario")

	if !errors.Is(err, negocio.ErrNaoAutorizado) {
		t.Fatalf("erro esperado de admin, recebido %v", err)
	}
}

func TestObterDashboardUsuarioNaoExigeNegocioId(t *testing.T) {
	repository := &dashboardRepositoryFake{resumoUsuario: &domain.ResumoUsuario{UsuarioId: "usuario-1", NegociosVinculados: 2, LandingPagesCriadas: 3}}
	useCase := NewUseCase(repository, &negocioRepositoryFake{})

	output, err := useCase.ObterDashboard(context.Background(), domain.FiltroPeriodo{UsuarioId: "outro-usuario", NegocioId: "negocio-1"}, "usuario-1", "usuario")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	dashboardUsuario, ok := output.(*dto.DashboardUsuarioOutput)
	if !ok || dashboardUsuario.NegociosVinculados != 2 || dashboardUsuario.LandingPagesCriadas != 3 {
		t.Fatalf("dashboard de usuario inesperado: %+v", output)
	}
	if repository.filtroUsuario.UsuarioId != "" || repository.filtroUsuario.NegocioId != "negocio-1" {
		t.Fatalf("filtro de usuario comum inesperado: %+v", repository.filtroUsuario)
	}
}

func TestObterDashboardNegocioMapeiaLandingPageMaisUsada(t *testing.T) {
	useCase := NewUseCase(&dashboardRepositoryFake{resumoNegocio: &domain.ResumoNegocio{NegocioId: "negocio-1", LandingPageMaisUsada: &domain.LandingPageUtilizada{Id: "lp-1", Nome: "LP", TotalExportacoes: 3}}}, &negocioRepositoryFake{})

	output, err := useCase.ObterDashboardNegocio(context.Background(), "negocio-1", domain.FiltroPeriodo{}, "usuario-1", "admin")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.LandingPageMaisUsada == nil || output.LandingPageMaisUsada.TotalExportacoes != 3 {
		t.Fatalf("landing page mais usada nao mapeada: %+v", output.LandingPageMaisUsada)
	}
}
