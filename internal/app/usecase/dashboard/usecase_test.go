package dashboard

import (
	"context"
	"errors"
	"testing"

	"nous/internal/app/dto"
	"nous/internal/domain/cliente"
	domain "nous/internal/domain/dashboard"
)

type dashboardRepositoryFake struct {
	resumoCliente *domain.ResumoCliente
	resumoUsuario *domain.ResumoUsuario
	resumoAdmin   *domain.ResumoAdmin
	filtroUsuario domain.FiltroPeriodo
}

func (r *dashboardRepositoryFake) ObterResumoCliente(context.Context, string, domain.FiltroPeriodo) (*domain.ResumoCliente, error) {
	return r.resumoCliente, nil
}

func (r *dashboardRepositoryFake) ObterResumoUsuario(_ context.Context, _ string, filtro domain.FiltroPeriodo) (*domain.ResumoUsuario, error) {
	r.filtroUsuario = filtro
	return r.resumoUsuario, nil
}

func (r *dashboardRepositoryFake) ObterResumoAdmin(context.Context, domain.FiltroPeriodo) (*domain.ResumoAdmin, error) {
	return r.resumoAdmin, nil
}

type clienteRepositoryFake struct{ err error }

func (r *clienteRepositoryFake) Salvar(context.Context, *cliente.Cliente) error { return nil }
func (r *clienteRepositoryFake) SalvarComVinculo(context.Context, *cliente.Cliente, *cliente.UsuarioCliente) error {
	return nil
}
func (r *clienteRepositoryFake) BuscarPorID(context.Context, string) (*cliente.Cliente, error) {
	return nil, cliente.ErrNaoEncontrado
}
func (r *clienteRepositoryFake) Atualizar(context.Context, *cliente.Cliente) error { return nil }
func (r *clienteRepositoryFake) SoftDelete(context.Context, string) error          { return nil }
func (r *clienteRepositoryFake) Listar(context.Context, cliente.FiltroListagem) ([]*cliente.Cliente, int, error) {
	return nil, 0, nil
}
func (r *clienteRepositoryFake) VincularUsuario(context.Context, *cliente.UsuarioCliente) error {
	return nil
}
func (r *clienteRepositoryFake) ListarPorUsuario(context.Context, string) ([]*cliente.Cliente, error) {
	return nil, nil
}
func (r *clienteRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return r.err
}

func TestObterDashboardClienteExigeClienteId(t *testing.T) {
	useCase := NewUseCase(&dashboardRepositoryFake{}, &clienteRepositoryFake{})

	_, err := useCase.ObterDashboardCliente(context.Background(), "", domain.FiltroPeriodo{}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrClienteObrigatorio) {
		t.Fatalf("erro esperado de cliente obrigatorio, recebido %v", err)
	}
}

func TestObterDashboardClienteValidaAcessoUsuario(t *testing.T) {
	useCase := NewUseCase(&dashboardRepositoryFake{resumoCliente: &domain.ResumoCliente{ClienteId: "cliente-1"}}, &clienteRepositoryFake{err: cliente.ErrNaoAutorizado})

	_, err := useCase.ObterDashboardCliente(context.Background(), "cliente-1", domain.FiltroPeriodo{}, "usuario-1", "usuario")

	if err == nil {
		t.Fatal("erro esperado de acesso ao cliente")
	}
}

func TestObterDashboardAdminExigeRoleAdmin(t *testing.T) {
	useCase := NewUseCase(&dashboardRepositoryFake{resumoAdmin: &domain.ResumoAdmin{}}, &clienteRepositoryFake{})

	_, err := useCase.ObterDashboardAdmin(context.Background(), domain.FiltroPeriodo{}, "usuario")

	if !errors.Is(err, cliente.ErrNaoAutorizado) {
		t.Fatalf("erro esperado de admin, recebido %v", err)
	}
}

func TestObterDashboardUsuarioNaoExigeClienteId(t *testing.T) {
	repository := &dashboardRepositoryFake{resumoUsuario: &domain.ResumoUsuario{UsuarioId: "usuario-1", ClientesVinculados: 2, LandingPagesCriadas: 3}}
	useCase := NewUseCase(repository, &clienteRepositoryFake{})

	output, err := useCase.ObterDashboard(context.Background(), domain.FiltroPeriodo{UsuarioId: "outro-usuario", ClienteId: "cliente-1"}, "usuario-1", "usuario")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	dashboardUsuario, ok := output.(*dto.DashboardUsuarioOutput)
	if !ok || dashboardUsuario.ClientesVinculados != 2 || dashboardUsuario.LandingPagesCriadas != 3 {
		t.Fatalf("dashboard de usuario inesperado: %+v", output)
	}
	if repository.filtroUsuario.UsuarioId != "" || repository.filtroUsuario.ClienteId != "cliente-1" {
		t.Fatalf("filtro de usuario comum inesperado: %+v", repository.filtroUsuario)
	}
}

func TestObterDashboardClienteMapeiaLandingPageMaisUsada(t *testing.T) {
	useCase := NewUseCase(&dashboardRepositoryFake{resumoCliente: &domain.ResumoCliente{ClienteId: "cliente-1", LandingPageMaisUsada: &domain.LandingPageUtilizada{Id: "lp-1", Nome: "LP", TotalExportacoes: 3}}}, &clienteRepositoryFake{})

	output, err := useCase.ObterDashboardCliente(context.Background(), "cliente-1", domain.FiltroPeriodo{}, "usuario-1", "admin")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.LandingPageMaisUsada == nil || output.LandingPageMaisUsada.TotalExportacoes != 3 {
		t.Fatalf("landing page mais usada nao mapeada: %+v", output.LandingPageMaisUsada)
	}
}
