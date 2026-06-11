package template

import (
	"context"
	"errors"
	"testing"

	"nous/internal/app/dto"
	"nous/internal/domain/cliente"
	"nous/internal/domain/plano"
	domain "nous/internal/domain/template"
)

type templateRepositoryFake struct {
	salvo     *domain.Template
	filtro    domain.FiltroListagem
	buscado   *domain.Template
	buscarErr error
}

func (r *templateRepositoryFake) Salvar(_ context.Context, t *domain.Template) error {
	t.Id = "template-1"
	r.salvo = t
	return nil
}

func (r *templateRepositoryFake) BuscarPorID(context.Context, string) (*domain.Template, error) {
	if r.buscado != nil {
		return r.buscado, nil
	}
	if r.buscarErr != nil {
		return nil, r.buscarErr
	}
	return nil, domain.ErrNaoEncontrado
}
func (r *templateRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *templateRepositoryFake) Listar(_ context.Context, filtro domain.FiltroListagem) ([]*domain.Template, int, error) {
	r.filtro = filtro
	return []*domain.Template{{Id: "template-1", Nome: "Template", Origem: domain.OrigemCliente, Tipo: domain.TipoWhiteLabel}}, 1, nil
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

func TestCriarTemplateAceitaArquivoSemConteudo(t *testing.T) {
	repository := &templateRepositoryFake{}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &planoRepositoryFake{})

	output, err := useCase.CriarTemplate(context.Background(), dto.CriarTemplateInput{
		Origem:       string(domain.OrigemSognix),
		Nome:         "White Label",
		Arquivo:      []byte("PK\x03\x04 conteudo zip"),
		NomeArquivo:  "site.zip",
		ContentType:  "application/zip",
		TamanhoBytes: 20,
	}, "admin-1", "admin")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !output.TemArquivo || output.NomeArquivo != "site.zip" || output.TamanhoBytes != 20 {
		t.Fatalf("metadados do arquivo nao propagados: %+v", output)
	}
	if repository.salvo == nil || len(repository.salvo.Arquivo) == 0 {
		t.Fatalf("arquivo binario nao foi salvo: %+v", repository.salvo)
	}
}

func TestCriarTemplateExigeConteudoOuArquivo(t *testing.T) {
	repository := &templateRepositoryFake{}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &planoRepositoryFake{})

	_, err := useCase.CriarTemplate(context.Background(), dto.CriarTemplateInput{
		Origem: string(domain.OrigemSognix),
		Nome:   "Vazio",
	}, "admin-1", "admin")

	if err == nil {
		t.Fatal("esperado erro quando nem conteudo nem arquivo sao informados")
	}
	if repository.salvo != nil {
		t.Fatal("template sem conteudo/arquivo nao deveria ser salvo")
	}
}

func TestBaixarTemplateSognixRetornaBytes(t *testing.T) {
	repository := &templateRepositoryFake{buscado: &domain.Template{
		Id:           "template-1",
		Origem:       domain.OrigemSognix,
		Arquivo:      []byte("zip-bytes"),
		NomeArquivo:  "site.zip",
		TamanhoBytes: 9,
	}}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &planoRepositoryFake{})

	dados, meta, err := useCase.BaixarTemplate(context.Background(), "template-1", "usuario-1", "usuario")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if string(dados) != "zip-bytes" || meta.NomeArquivo != "site.zip" {
		t.Fatalf("download inesperado: dados=%q meta=%+v", dados, meta)
	}
}

func TestBaixarTemplateClienteExigeAcesso(t *testing.T) {
	repository := &templateRepositoryFake{buscado: &domain.Template{
		Id:        "template-1",
		Origem:    domain.OrigemCliente,
		ClienteId: "cliente-1",
		Arquivo:   []byte("zip-bytes"),
	}}
	useCase := NewUseCase(repository, &clienteRepositoryFake{err: cliente.ErrNaoEncontrado}, &planoRepositoryFake{})

	_, _, err := useCase.BaixarTemplate(context.Background(), "template-1", "usuario-1", "usuario")
	if err == nil {
		t.Fatal("esperado erro de acesso ao baixar template de outro cliente")
	}
}

func TestBaixarTemplateSemArquivo(t *testing.T) {
	repository := &templateRepositoryFake{buscado: &domain.Template{
		Id:       "template-1",
		Origem:   domain.OrigemSognix,
		Conteudo: "<html></html>",
	}}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, &planoRepositoryFake{})

	_, _, err := useCase.BaixarTemplate(context.Background(), "template-1", "admin-1", "admin")
	if !errors.Is(err, domain.ErrSemArquivo) {
		t.Fatalf("esperado ErrSemArquivo, recebido %v", err)
	}
}

func TestListarTemplatesUsuarioExigeCliente(t *testing.T) {
	useCase := NewUseCase(&templateRepositoryFake{}, &clienteRepositoryFake{}, &planoRepositoryFake{})

	_, err := useCase.ListarTemplates(context.Background(), domain.FiltroListagem{}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrClienteObrigatorio) {
		t.Fatalf("erro esperado cliente obrigatorio, recebido %v", err)
	}
}
