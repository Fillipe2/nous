package landingpage

import (
	"context"
	"errors"
	"testing"
	"time"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/evento"
	domain "buildnix-backend/internal/domain/landingpage"
	"buildnix-backend/internal/domain/plano"
	"buildnix-backend/internal/domain/template"
)

type landingPageRepositoryFake struct {
	landingPage       *domain.LandingPage
	salva             *domain.LandingPage
	totalPorCliente   int
	contarChamadas    int
	autosaveChamadas  int
	atualizarLockPara *string
}

func (r *landingPageRepositoryFake) Salvar(_ context.Context, lp *domain.LandingPage) error {
	lp.Id = "lp-nova"
	r.salva = lp
	return nil
}

func (r *landingPageRepositoryFake) BuscarPorID(context.Context, string) (*domain.LandingPage, error) {
	if r.landingPage == nil {
		return nil, domain.ErrNaoEncontrada
	}
	return r.landingPage, nil
}

func (r *landingPageRepositoryFake) Atualizar(context.Context, *domain.LandingPage) error { return nil }
func (r *landingPageRepositoryFake) SoftDelete(context.Context, string) error             { return nil }
func (r *landingPageRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.LandingPage, int, error) {
	return nil, 0, nil
}
func (r *landingPageRepositoryFake) ContarPorCliente(context.Context, string) (int, error) {
	r.contarChamadas++
	return r.totalPorCliente, nil
}
func (r *landingPageRepositoryFake) AtualizarAutosave(context.Context, string, []byte, int) (int, error) {
	r.autosaveChamadas++
	return 2, nil
}
func (r *landingPageRepositoryFake) AtualizarLock(_ context.Context, _ string, usuarioId *string) error {
	r.atualizarLockPara = usuarioId
	return nil
}
func (r *landingPageRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return nil
}
func (r *landingPageRepositoryFake) AtualizarPublicacao(context.Context, string, string, string, time.Time) error {
	return nil
}
func (r *landingPageRepositoryFake) AtualizarDespublicacao(context.Context, string) error {
	return nil
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

type templateRepositoryFake struct{ template *template.Template }

func (r *templateRepositoryFake) Salvar(context.Context, *template.Template) error { return nil }
func (r *templateRepositoryFake) BuscarPorID(context.Context, string) (*template.Template, error) {
	if r.template == nil {
		return nil, template.ErrNaoEncontrado
	}
	return r.template, nil
}
func (r *templateRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *templateRepositoryFake) Listar(context.Context, template.FiltroListagem) ([]*template.Template, int, error) {
	return nil, 0, nil
}

type publisherFake struct{}

func (publisherFake) Publicar(context.Context, evento.Evento) error { return nil }

func TestCriarLandingPageCustomizavelRespeitaLimiteDoPlano(t *testing.T) {
	repository := &landingPageRepositoryFake{totalPorCliente: 1}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, publisherFake{}, &planoRepositoryFake{plano: &plano.Plano{Limites: plano.Limites{LandingPagesCustomizaveis: true, MaxLandingPages: 1}}}, &templateRepositoryFake{}, nil, nil)

	_, err := useCase.CriarLandingPage(context.Background(), dto.CriarLandingPageInput{ClienteId: "cliente-1", Nome: "LP", Tipo: string(domain.TipoCustomizavel), Conteudo: "html"}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrPlanoNaoPermite) {
		t.Fatalf("erro esperado de limite de plano, recebido %v", err)
	}
	if repository.contarChamadas != 1 {
		t.Fatalf("limite deveria usar ContarPorCliente uma vez, chamadas=%d", repository.contarChamadas)
	}
	if repository.salva != nil {
		t.Fatal("landing page acima do limite nao deveria ser salva")
	}
}

func TestCriarLandingPageWhiteLabelExigeTemplate(t *testing.T) {
	repository := &landingPageRepositoryFake{}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, publisherFake{}, &planoRepositoryFake{plano: &plano.Plano{}}, &templateRepositoryFake{}, nil, nil)

	_, err := useCase.CriarLandingPage(context.Background(), dto.CriarLandingPageInput{ClienteId: "cliente-1", Nome: "LP", Tipo: string(domain.TipoWhiteLabel)}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrTemplateObrigatorio) {
		t.Fatalf("erro esperado de template obrigatorio, recebido %v", err)
	}
	if repository.salva != nil {
		t.Fatal("white label sem template nao deveria ser salva")
	}
}

func TestCriarLandingPageWhiteLabelUsaTemplateInicial(t *testing.T) {
	repository := &landingPageRepositoryFake{}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, publisherFake{}, &planoRepositoryFake{plano: &plano.Plano{}}, &templateRepositoryFake{template: &template.Template{Id: "template-1", ClienteId: "cliente-1", CategoriaId: "categoria-1", Tipo: template.TipoWhiteLabel, Origem: template.OrigemCliente, Conteudo: "html-template"}}, nil, nil)

	output, err := useCase.CriarLandingPage(context.Background(), dto.CriarLandingPageInput{ClienteId: "cliente-1", Nome: "LP", Tipo: string(domain.TipoWhiteLabel), TemplateId: "template-1"}, "usuario-1", "usuario")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.TemplateId != "template-1" || output.CategoriaId != "categoria-1" || output.Conteudo != "html-template" {
		t.Fatalf("template inicial nao aplicado: %+v", output)
	}
	if repository.salva == nil || repository.salva.IsDraft != true || repository.salva.Status != domain.StatusRascunho {
		t.Fatalf("landing page salva inesperada: %+v", repository.salva)
	}
}

func TestAutosaveRecusaLandingPageWhiteLabel(t *testing.T) {
	repository := &landingPageRepositoryFake{landingPage: &domain.LandingPage{Id: "lp-1", ClienteId: "cliente-1", Tipo: domain.TipoWhiteLabel, Version: 1}}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, publisherFake{}, &planoRepositoryFake{}, &templateRepositoryFake{}, nil, nil)

	_, err := useCase.Autosave(context.Background(), "lp-1", "usuario-1", "usuario", dto.AutosaveInput{AutosaveData: []byte(`{"a":1}`), Version: 1})

	if !errors.Is(err, domain.ErrPlanoNaoPermite) {
		t.Fatalf("erro esperado de plano, recebido %v", err)
	}
	if repository.autosaveChamadas != 0 {
		t.Fatal("autosave white label nao deveria persistir rascunho")
	}
}

func TestBloquearEdicaoRespeitaLockAtivoDeOutroUsuario(t *testing.T) {
	outroUsuario := "usuario-2"
	lockedAt := time.Now()
	repository := &landingPageRepositoryFake{landingPage: &domain.LandingPage{Id: "lp-1", ClienteId: "cliente-1", Tipo: domain.TipoCustomizavel, LockedBy: &outroUsuario, LockedAt: &lockedAt}}
	useCase := NewUseCase(repository, &clienteRepositoryFake{}, publisherFake{}, &planoRepositoryFake{}, &templateRepositoryFake{}, nil, nil)

	_, err := useCase.BloquearEdicao(context.Background(), "lp-1", "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrBloqueada) {
		t.Fatalf("erro esperado de lock ativo, recebido %v", err)
	}
	if repository.atualizarLockPara != nil {
		t.Fatal("lock ativo de outro usuario nao deveria ser sobrescrito")
	}
}
