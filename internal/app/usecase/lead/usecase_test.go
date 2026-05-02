package lead

import (
	"context"
	"errors"
	"testing"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/evento"
	domain "buildnix-backend/internal/domain/lead"
)

type leadRepositoryFake struct {
	salvo *domain.Lead
}

func (r *leadRepositoryFake) Salvar(_ context.Context, l *domain.Lead) error {
	l.Id = "lead-test"
	r.salvo = l
	return nil
}

func (r *leadRepositoryFake) BuscarPorID(context.Context, string) (*domain.Lead, error) {
	return nil, domain.ErrNaoEncontrado
}

func (r *leadRepositoryFake) AtualizarStatus(context.Context, string, domain.Status) error {
	return nil
}

func (r *leadRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.Lead, int, error) {
	return nil, 0, nil
}

type publisherFake struct {
	err       error
	publicado bool
}

func (p *publisherFake) Publicar(context.Context, evento.Evento) error {
	p.publicado = true
	return p.err
}

func TestCapturarLeadRejeitaEmailInvalido(t *testing.T) {
	repository := &leadRepositoryFake{}
	useCase := NewUseCase(repository, &publisherFake{})

	_, err := useCase.CapturarLead(context.Background(), dto.CapturarLeadInput{
		LandingPageId: "lp-test",
		Nome:          "Lead Teste",
		Email:         "lead@",
		Telefone:      "11999999999",
	}, "127.0.0.1", "test")

	if err == nil {
		t.Fatal("erro esperado para email invalido")
	}
	if repository.salvo != nil {
		t.Fatal("lead com email invalido nao deveria ser salvo")
	}
}

func TestCapturarLeadNormalizaEmailEIgnoraErroDePublicacao(t *testing.T) {
	repository := &leadRepositoryFake{}
	publisher := &publisherFake{err: errors.New("observer indisponivel")}
	useCase := NewUseCase(repository, publisher)

	output, err := useCase.CapturarLead(context.Background(), dto.CapturarLeadInput{
		LandingPageId: "lp-test",
		Nome:          "Lead Teste",
		Email:         " lead@sognix.com.br ",
		Whatsapp:      "11999999999",
	}, "127.0.0.1", "test")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Email != "lead@sognix.com.br" || repository.salvo.Email != "lead@sognix.com.br" {
		t.Fatal("email nao foi normalizado antes de salvar")
	}
	if !publisher.publicado {
		t.Fatal("evento de lead deveria ter sido tentado")
	}
}
