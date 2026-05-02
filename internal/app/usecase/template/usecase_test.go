package template

import (
	"context"
	"errors"
	"testing"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/plano"
	domain "buildnix-backend/internal/domain/template"
)

type templateRepositoryFake struct {
	salvo  *domain.Template
	filtro domain.FiltroListagem
}

func (r *templateRepositoryFake) Salvar(_ context.Context, t *domain.Template) error {
	t.Id = "template-1"
	r.salvo = t
	return nil
}

func (r *templateRepositoryFake) BuscarPorID(context.Context, string) (*domain.Template, error) {
	return nil, domain.ErrNaoEncontrado
}
func (r *templateRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *templateRepositoryFake) Listar(_ context.Context, filtro domain.FiltroListagem) ([]*domain.Template, int, error) {
	r.filtro = filtro
	return []*domain.Template{{Id: "template-1", Nome: "Template", Origem: domain.OrigemCliente, Tipo: domain.TipoWhiteLabel}}, 1, nil
}

type clienteRepositoryFake struct{ err error }

func (r *clienteRepositoryFake) Salvar(context.Context, *cliente.Cliente) error { return nil }
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

type planoRepositoryFake struct{ plano *plano.Plano }

func (r *planoRepositoryFake) BuscarPlanoAtivoPorCliente(context.Context, string) (*plano.Plano, error) {
	if r.plano == nil {
		return nil, plano.ErrAssinaturaAtivaNaoEncontrada
	}
	return r.plano, nil
}

func TestCriarTemplateSognixExigeAdmin(t *testing.T) {
	repository := &templateRepositoryFake{}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &planoRepositoryFake{})

	_, err := useCase.CriarTemplate(context.Background(), dto.CriarTemplateInput{Origem: string(domain.OrigemSognix), Nome: "Template", Conteudo: "html"}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrNaoAutorizado) {
		t.Fatalf("erro esperado de autorizacao, recebido %v", err)
	}
	if repository.salvo != nil {
		t.Fatal("template sognix sem admin nao deveria ser salvo")
	}
}

func TestCriarTemplateClienteBaseCustomizavelExigePlano(t *testing.T) {
	repository := &templateRepositoryFake{}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &planoRepositoryFake{plano: &plano.Plano{Limites: plano.Limites{LandingPagesCustomizaveis: false}}})

	_, err := useCase.CriarTemplate(context.Background(), dto.CriarTemplateInput{ClienteId: "cliente-1", Origem: string(domain.OrigemCliente), Tipo: string(domain.TipoBaseCustomizavel), Nome: "Template", Conteudo: "html"}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrPlanoNaoPermite) {
		t.Fatalf("erro esperado de plano, recebido %v", err)
	}
	if repository.salvo != nil {
		t.Fatal("template sem plano customizavel nao deveria ser salvo")
	}
}

func TestCriarTemplateClienteBaseCustomizavelPermitido(t *testing.T) {
	repository := &templateRepositoryFake{}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &planoRepositoryFake{plano: &plano.Plano{Limites: plano.Limites{LandingPagesCustomizaveis: true}}})

	output, err := useCase.CriarTemplate(context.Background(), dto.CriarTemplateInput{ClienteId: "cliente-1", Origem: string(domain.OrigemCliente), Tipo: string(domain.TipoBaseCustomizavel), Nome: "Template", Conteudo: "html"}, "usuario-1", "usuario")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Id != "template-1" || output.Tipo != string(domain.TipoBaseCustomizavel) || repository.salvo.ClienteId != "cliente-1" {
		t.Fatalf("template salvo inesperado: output=%+v salvo=%+v", output, repository.salvo)
	}
}

func TestListarTemplatesUsuarioExigeCliente(t *testing.T) {
	useCase := NewUseCase(&templateRepositoryFake{}, &clienteRepositoryFake{}, &planoRepositoryFake{})

	_, err := useCase.ListarTemplates(context.Background(), domain.FiltroListagem{}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrClienteObrigatorio) {
		t.Fatalf("erro esperado cliente obrigatorio, recebido %v", err)
	}
}
