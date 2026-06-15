package negocio

import (
	"context"
	"errors"
	"testing"

	"nous/internal/app/dto"
	domain "nous/internal/domain/negocio"
)

type negocioRepositoryFake struct {
	negocio    *domain.Negocio
	salvo      *domain.Negocio
	atualizado *domain.Negocio
	vinculo    *domain.UsuarioNegocio
}

func (r *negocioRepositoryFake) Salvar(_ context.Context, c *domain.Negocio) error {
	c.Id = "negocio-novo"
	r.salvo = c
	return nil
}

func (r *negocioRepositoryFake) SalvarComVinculo(_ context.Context, c *domain.Negocio, uc *domain.UsuarioNegocio) error {
	c.Id = "negocio-novo"
	uc.NegocioId = c.Id
	r.salvo = c
	r.vinculo = uc
	return nil
}

func (r *negocioRepositoryFake) BuscarPorID(context.Context, string) (*domain.Negocio, error) {
	if r.negocio == nil {
		return nil, domain.ErrNaoEncontrado
	}
	return r.negocio, nil
}

func (r *negocioRepositoryFake) Atualizar(_ context.Context, c *domain.Negocio) error {
	r.atualizado = c
	return nil
}

func (r *negocioRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *negocioRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.Negocio, int, error) {
	return nil, 0, nil
}
func (r *negocioRepositoryFake) VincularUsuario(_ context.Context, uc *domain.UsuarioNegocio) error {
	r.vinculo = uc
	return nil
}
func (r *negocioRepositoryFake) ListarPorUsuario(context.Context, string) ([]*domain.Negocio, error) {
	return nil, nil
}
func (r *negocioRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return nil
}

func TestCriarNegocioValidaNomeObrigatorio(t *testing.T) {
	useCase := NewUseCase(&negocioRepositoryFake{})

	_, err := useCase.CriarNegocio(context.Background(), dto.CriarNegocioInput{}, "usuario-1")

	if err == nil {
		t.Fatal("erro esperado para nome vazio")
	}
}

func TestCriarNegocioDefineStatusAtivoEVersaoInicial(t *testing.T) {
	repository := &negocioRepositoryFake{}
	useCase := NewUseCase(repository)

	output, err := useCase.CriarNegocio(context.Background(), dto.CriarNegocioInput{Nome: "Negocio", Email: "negocio@sognix.com.br", Telefone: "1130000000", Whatsapp: "11999999999"}, "usuario-1")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Id != "negocio-novo" || output.Status != string(domain.StatusAtivo) || output.Version != 1 || output.Telefone != "1130000000" || output.Whatsapp != "11999999999" {
		t.Fatalf("saida inesperada: %+v", output)
	}
	if repository.salvo == nil || repository.salvo.Status != domain.StatusAtivo || repository.salvo.Version != 1 {
		t.Fatalf("negocio salvo inesperado: %+v", repository.salvo)
	}
	if repository.vinculo == nil || repository.vinculo.UsuarioId != "usuario-1" || repository.vinculo.NegocioId != "negocio-novo" {
		t.Fatalf("vinculo automatico inesperado: %+v", repository.vinculo)
	}
}

func TestAtualizarNegocioDetectaConflitoDeVersao(t *testing.T) {
	repository := &negocioRepositoryFake{negocio: &domain.Negocio{Id: "negocio-1", Version: 2}}
	useCase := NewUseCase(repository)

	_, err := useCase.AtualizarNegocio(context.Background(), "negocio-1", dto.AtualizarNegocioInput{Nome: "Novo", Version: 1})

	if !errors.Is(err, domain.ErrVersaoConflito) {
		t.Fatalf("erro esperado de versao, recebido %v", err)
	}
	if repository.atualizado != nil {
		t.Fatal("negocio com conflito nao deveria ser atualizado")
	}
}

func TestVincularUsuarioRegistraRelacionamento(t *testing.T) {
	repository := &negocioRepositoryFake{}
	useCase := NewUseCase(repository)

	if err := useCase.VincularUsuario(context.Background(), dto.VincularUsuarioNegocioInput{UsuarioId: "usuario-1", NegocioId: "negocio-1"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repository.vinculo == nil || repository.vinculo.UsuarioId != "usuario-1" || repository.vinculo.NegocioId != "negocio-1" {
		t.Fatalf("vinculo inesperado: %+v", repository.vinculo)
	}
}
