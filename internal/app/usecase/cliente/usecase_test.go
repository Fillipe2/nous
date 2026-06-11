package cliente

import (
	"context"
	"errors"
	"testing"

	"nous/internal/app/dto"
	domain "nous/internal/domain/cliente"
)

type clienteRepositoryFake struct {
	cliente    *domain.Cliente
	salvo      *domain.Cliente
	atualizado *domain.Cliente
	vinculo    *domain.UsuarioCliente
}

func (r *clienteRepositoryFake) Salvar(_ context.Context, c *domain.Cliente) error {
	c.Id = "cliente-novo"
	r.salvo = c
	return nil
}

func (r *clienteRepositoryFake) SalvarComVinculo(_ context.Context, c *domain.Cliente, uc *domain.UsuarioCliente) error {
	c.Id = "cliente-novo"
	uc.ClienteId = c.Id
	r.salvo = c
	r.vinculo = uc
	return nil
}

func (r *clienteRepositoryFake) BuscarPorID(context.Context, string) (*domain.Cliente, error) {
	if r.cliente == nil {
		return nil, domain.ErrNaoEncontrado
	}
	return r.cliente, nil
}

func (r *clienteRepositoryFake) Atualizar(_ context.Context, c *domain.Cliente) error {
	r.atualizado = c
	return nil
}

func (r *clienteRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *clienteRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.Cliente, int, error) {
	return nil, 0, nil
}
func (r *clienteRepositoryFake) VincularUsuario(_ context.Context, uc *domain.UsuarioCliente) error {
	r.vinculo = uc
	return nil
}
func (r *clienteRepositoryFake) ListarPorUsuario(context.Context, string) ([]*domain.Cliente, error) {
	return nil, nil
}
func (r *clienteRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return nil
}

func TestCriarClienteValidaNomeObrigatorio(t *testing.T) {
	useCase := NewUseCase(&clienteRepositoryFake{})

	_, err := useCase.CriarCliente(context.Background(), dto.CriarClienteInput{}, "usuario-1")

	if err == nil {
		t.Fatal("erro esperado para nome vazio")
	}
}

func TestCriarClienteDefineStatusAtivoEVersaoInicial(t *testing.T) {
	repository := &clienteRepositoryFake{}
	useCase := NewUseCase(repository)

	output, err := useCase.CriarCliente(context.Background(), dto.CriarClienteInput{Nome: "Cliente", Email: "cliente@sognix.com.br", Telefone: "1130000000", Whatsapp: "11999999999"}, "usuario-1")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Id != "cliente-novo" || output.Status != string(domain.StatusAtivo) || output.Version != 1 || output.Telefone != "1130000000" || output.Whatsapp != "11999999999" {
		t.Fatalf("saida inesperada: %+v", output)
	}
	if repository.salvo == nil || repository.salvo.Status != domain.StatusAtivo || repository.salvo.Version != 1 {
		t.Fatalf("cliente salvo inesperado: %+v", repository.salvo)
	}
	if repository.vinculo == nil || repository.vinculo.UsuarioId != "usuario-1" || repository.vinculo.ClienteId != "cliente-novo" {
		t.Fatalf("vinculo automatico inesperado: %+v", repository.vinculo)
	}
}

func TestAtualizarClienteDetectaConflitoDeVersao(t *testing.T) {
	repository := &clienteRepositoryFake{cliente: &domain.Cliente{Id: "cliente-1", Version: 2}}
	useCase := NewUseCase(repository)

	_, err := useCase.AtualizarCliente(context.Background(), "cliente-1", dto.AtualizarClienteInput{Nome: "Novo", Version: 1})

	if !errors.Is(err, domain.ErrVersaoConflito) {
		t.Fatalf("erro esperado de versao, recebido %v", err)
	}
	if repository.atualizado != nil {
		t.Fatal("cliente com conflito nao deveria ser atualizado")
	}
}

func TestVincularUsuarioRegistraRelacionamento(t *testing.T) {
	repository := &clienteRepositoryFake{}
	useCase := NewUseCase(repository)

	if err := useCase.VincularUsuario(context.Background(), dto.VincularUsuarioClienteInput{UsuarioId: "usuario-1", ClienteId: "cliente-1"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repository.vinculo == nil || repository.vinculo.UsuarioId != "usuario-1" || repository.vinculo.ClienteId != "cliente-1" {
		t.Fatalf("vinculo inesperado: %+v", repository.vinculo)
	}
}
