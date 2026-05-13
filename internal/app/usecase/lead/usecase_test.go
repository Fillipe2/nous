package lead

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/evento"
	"buildnix-backend/internal/domain/landingpage"
	domain "buildnix-backend/internal/domain/lead"
)

type leadRepositoryFake struct {
	salvo                *domain.Lead
	duplicado            *domain.Lead
	statusAtualizacao    domain.AtualizacaoStatus
	statusID             string
	anonimizadoID        string
	anonimizadoExpirados int
}

func (r *leadRepositoryFake) Salvar(_ context.Context, l *domain.Lead) error {
	l.Id = "lead-test"
	r.salvo = l
	return nil
}

func (r *leadRepositoryFake) BuscarPorID(context.Context, string) (*domain.Lead, error) {
	return nil, domain.ErrNaoEncontrado
}

func (r *leadRepositoryFake) BuscarDuplicado(context.Context, string, string, time.Time) (*domain.Lead, error) {
	if r.duplicado == nil {
		return nil, domain.ErrNaoEncontrado
	}
	return r.duplicado, nil
}

func (r *leadRepositoryFake) AtualizarStatus(_ context.Context, id string, atualizacao domain.AtualizacaoStatus, _ time.Time) error {
	r.statusID = id
	r.statusAtualizacao = atualizacao
	return nil
}

func (r *leadRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.Lead, int, error) {
	return nil, 0, nil
}

func (r *leadRepositoryFake) Anonimizar(_ context.Context, id string, _ time.Time) error {
	r.anonimizadoID = id
	return nil
}

func (r *leadRepositoryFake) AnonimizarExpirados(context.Context, time.Time) (int, error) {
	r.anonimizadoExpirados++
	return 3, nil
}

type publisherFake struct {
	err       error
	publicado bool
	ultimo    evento.Evento
}

func (p *publisherFake) Publicar(_ context.Context, e evento.Evento) error {
	p.publicado = true
	p.ultimo = e
	return p.err
}

type landingPageFake struct {
	lp  *landingpage.LandingPage
	err error
}

func (f *landingPageFake) Salvar(context.Context, *landingpage.LandingPage) error { return nil }
func (f *landingPageFake) BuscarPorID(context.Context, string) (*landingpage.LandingPage, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.lp, nil
}
func (f *landingPageFake) Atualizar(context.Context, *landingpage.LandingPage) error { return nil }
func (f *landingPageFake) SoftDelete(context.Context, string) error                  { return nil }
func (f *landingPageFake) Listar(context.Context, landingpage.FiltroListagem) ([]*landingpage.LandingPage, int, error) {
	return nil, 0, nil
}
func (f *landingPageFake) ContarPorCliente(context.Context, string) (int, error) { return 0, nil }
func (f *landingPageFake) AtualizarAutosave(context.Context, string, []byte, int) (int, error) {
	return 0, nil
}
func (f *landingPageFake) AtualizarLock(context.Context, string, *string) error { return nil }
func (f *landingPageFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return nil
}
func (f *landingPageFake) AtualizarPublicacao(context.Context, string, string, string, time.Time) error {
	return nil
}
func (f *landingPageFake) AtualizarDespublicacao(context.Context, string) error { return nil }

func lpPublicada() *landingPageFake {
	return &landingPageFake{lp: &landingpage.LandingPage{Id: "lp-test", ClienteId: "cliente-1", Status: landingpage.StatusPublicada}}
}

func inputBase() dto.CapturarLeadInput {
	return dto.CapturarLeadInput{
		LandingPageId: "lp-test",
		Nome:          "Lead Teste",
		Email:         "lead@sognix.com.br",
		Telefone:      "11999999999",
		Consentimento: true,
	}
}

func TestCapturarLeadRejeitaEmailInvalido(t *testing.T) {
	repository := &leadRepositoryFake{}
	useCase := NewUseCase(repository, lpPublicada(), &publisherFake{})

	in := inputBase()
	in.Email = "lead@"
	_, err := useCase.CapturarLead(context.Background(), in, "127.0.0.1", "test")

	if err == nil {
		t.Fatal("erro esperado para email invalido")
	}
	if repository.salvo != nil {
		t.Fatal("lead com email invalido nao deveria ser salvo")
	}
}

func TestCapturarLeadRejeitaSemConsentimento(t *testing.T) {
	repository := &leadRepositoryFake{}
	useCase := NewUseCase(repository, lpPublicada(), &publisherFake{})

	in := inputBase()
	in.Consentimento = false
	_, err := useCase.CapturarLead(context.Background(), in, "127.0.0.1", "test")

	if !errors.Is(err, domain.ErrConsentimentoObrigatorio) {
		t.Fatalf("esperado ErrConsentimentoObrigatorio, recebido %v", err)
	}
}

func TestCapturarLeadRejeitaHoneypotPreenchido(t *testing.T) {
	repository := &leadRepositoryFake{}
	useCase := NewUseCase(repository, lpPublicada(), &publisherFake{})

	in := inputBase()
	in.Honeypot = "http://spam.example"
	_, err := useCase.CapturarLead(context.Background(), in, "127.0.0.1", "bot")

	if !errors.Is(err, domain.ErrSpamDetectado) {
		t.Fatalf("esperado ErrSpamDetectado, recebido %v", err)
	}
	if repository.salvo != nil {
		t.Fatal("submissao de spam nao deveria persistir")
	}
}

func TestCapturarLeadRejeitaLandingPageNaoPublicada(t *testing.T) {
	repository := &leadRepositoryFake{}
	lpFake := &landingPageFake{lp: &landingpage.LandingPage{Id: "lp-test", ClienteId: "cliente-1", Status: landingpage.StatusRascunho}}
	useCase := NewUseCase(repository, lpFake, &publisherFake{})

	_, err := useCase.CapturarLead(context.Background(), inputBase(), "127.0.0.1", "test")
	if !errors.Is(err, domain.ErrLandingPageNaoPublicada) {
		t.Fatalf("esperado ErrLandingPageNaoPublicada, recebido %v", err)
	}
}

func TestCapturarLeadDeduplicaJanelaCurta(t *testing.T) {
	existente := &domain.Lead{Id: "lead-existente", LandingPageId: "lp-test", Email: "lead@sognix.com.br", Status: domain.StatusNovo, CriadoEm: time.Now().UTC()}
	repository := &leadRepositoryFake{duplicado: existente}
	publisher := &publisherFake{}
	useCase := NewUseCase(repository, lpPublicada(), publisher)

	output, err := useCase.CapturarLead(context.Background(), inputBase(), "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Id != "lead-existente" {
		t.Fatalf("esperado retornar lead existente, recebeu %s", output.Id)
	}
	if repository.salvo != nil {
		t.Fatal("dedupe deveria ter evitado salvar")
	}
	if publisher.publicado {
		t.Fatal("dedupe deveria ter evitado publicar")
	}
}

func TestCapturarLeadEnriquecePayloadComUTMEClienteId(t *testing.T) {
	repository := &leadRepositoryFake{}
	publisher := &publisherFake{}
	useCase := NewUseCase(repository, lpPublicada(), publisher)

	in := inputBase()
	in.Email = " LEAD@sognix.com.br "
	in.Telefone = "(11) 99999-9999"
	in.UTM = &dto.UTMInput{Source: "google", Medium: "cpc", Campaign: "lancamento"}
	in.Referrer = "https://anuncio.example/path"

	output, err := useCase.CapturarLead(context.Background(), in, "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if output.Email != "lead@sognix.com.br" || repository.salvo.Email != "lead@sognix.com.br" {
		t.Fatalf("email nao normalizado: %q", repository.salvo.Email)
	}
	if repository.salvo.Telefone != "1199999999999" && repository.salvo.Telefone != "11999999999" {
		// telefone normalizado mantem apenas digitos
		if !strings.ContainsAny(repository.salvo.Telefone, "0123456789") || strings.ContainsAny(repository.salvo.Telefone, "() -") {
			t.Fatalf("telefone nao foi normalizado para apenas digitos: %q", repository.salvo.Telefone)
		}
	}
	if repository.salvo.ClienteId != "cliente-1" {
		t.Fatalf("cliente_id deveria ser derivado da landingpage, recebido %q", repository.salvo.ClienteId)
	}
	if !repository.salvo.Consentimento || repository.salvo.ConsentimentoEm == nil {
		t.Fatal("consentimento e timestamp deveriam estar preenchidos")
	}
	if !publisher.publicado {
		t.Fatal("evento de lead deveria ter sido publicado")
	}
	payload := publisher.ultimo.Payload
	if payload == nil {
		t.Fatal("payload nao deveria ser nil")
	}
	if payload["cliente_id"] != "cliente-1" {
		t.Fatalf("payload sem cliente_id correto: %+v", payload)
	}
	if payload["consentimento"] != true {
		t.Fatalf("payload sem consentimento: %+v", payload)
	}
	utm, ok := payload["utm"].(map[string]string)
	if !ok || utm["utm_source"] != "google" || utm["utm_campaign"] != "lancamento" {
		t.Fatalf("payload sem UTM correto: %+v", payload["utm"])
	}
}

func TestCapturarLeadIgnoraErroDePublicacao(t *testing.T) {
	repository := &leadRepositoryFake{}
	publisher := &publisherFake{err: errors.New("observer indisponivel")}
	useCase := NewUseCase(repository, lpPublicada(), publisher)

	output, err := useCase.CapturarLead(context.Background(), inputBase(), "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output == nil || repository.salvo == nil {
		t.Fatal("lead deveria ter sido salvo mesmo com erro de publicacao")
	}
}

func TestAtualizarStatusValidaEnumEMotivo(t *testing.T) {
	repository := &leadRepositoryFake{}
	useCase := NewUseCase(repository, lpPublicada(), &publisherFake{})

	if err := useCase.AtualizarStatusLead(context.Background(), "lead-1", dto.AtualizarStatusLeadInput{Status: "invalido"}); !errors.Is(err, domain.ErrStatusInvalido) {
		t.Fatalf("esperado ErrStatusInvalido, recebido %v", err)
	}
	if err := useCase.AtualizarStatusLead(context.Background(), "lead-1", dto.AtualizarStatusLeadInput{Status: string(domain.StatusPerdido)}); !errors.Is(err, domain.ErrMotivoObrigatorio) {
		t.Fatalf("esperado ErrMotivoObrigatorio, recebido %v", err)
	}
	if err := useCase.AtualizarStatusLead(context.Background(), "lead-1", dto.AtualizarStatusLeadInput{Status: string(domain.StatusPerdido), Motivo: "preco"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repository.statusAtualizacao.Status != domain.StatusPerdido || repository.statusAtualizacao.Motivo != "preco" {
		t.Fatalf("atualizacao de status nao propagou motivo: %+v", repository.statusAtualizacao)
	}
}
