package negocio

import (
	"context"
	"errors"
	"testing"

	"nous/internal/app/dto"
	"nous/internal/domain/cliente"
	domain "nous/internal/domain/negocio"
)

type negocioRepositoryFake struct {
	negocio    *domain.Negocio
	salvo      *domain.Negocio
	atualizado *domain.Negocio
}

func (r *negocioRepositoryFake) Salvar(_ context.Context, n *domain.Negocio) error {
	n.Id = "negocio-novo"
	r.salvo = n
	return nil
}

func (r *negocioRepositoryFake) BuscarPorID(context.Context, string) (*domain.Negocio, error) {
	if r.negocio == nil {
		return nil, domain.ErrNaoEncontrado
	}
	return r.negocio, nil
}

func (r *negocioRepositoryFake) Atualizar(_ context.Context, n *domain.Negocio) error {
	r.atualizado = n
	return nil
}

func (r *negocioRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *negocioRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.Negocio, int, error) {
	return nil, 0, nil
}

type clienteAcessoRepositoryFake struct{ err error }

func (r *clienteAcessoRepositoryFake) Salvar(context.Context, *cliente.Cliente) error { return nil }
func (r *clienteAcessoRepositoryFake) BuscarPorID(context.Context, string) (*cliente.Cliente, error) {
	return nil, cliente.ErrNaoEncontrado
}
func (r *clienteAcessoRepositoryFake) Atualizar(context.Context, *cliente.Cliente) error { return nil }
func (r *clienteAcessoRepositoryFake) SoftDelete(context.Context, string) error          { return nil }
func (r *clienteAcessoRepositoryFake) Listar(context.Context, cliente.FiltroListagem) ([]*cliente.Cliente, int, error) {
	return nil, 0, nil
}
func (r *clienteAcessoRepositoryFake) VincularUsuario(context.Context, *cliente.UsuarioCliente) error {
	return nil
}
func (r *clienteAcessoRepositoryFake) ListarPorUsuario(context.Context, string) ([]*cliente.Cliente, error) {
	return nil, nil
}
func (r *clienteAcessoRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return r.err
}

func TestCriarNegocioValidaCamposObrigatorios(t *testing.T) {
	useCase := NewUseCase(&negocioRepositoryFake{}, &clienteAcessoRepositoryFake{})

	_, err := useCase.CriarNegocio(context.Background(), dto.CriarNegocioInput{ClienteId: "cliente-1"}, "usuario-1", "usuario")

	if err == nil {
		t.Fatal("erro esperado para nome vazio")
	}
}

func TestCriarNegocioExigeAcessoAoCliente(t *testing.T) {
	repository := &negocioRepositoryFake{}
	useCase := NewUseCase(repository, &clienteAcessoRepositoryFake{err: cliente.ErrNaoAutorizado})

	_, err := useCase.CriarNegocio(context.Background(), dto.CriarNegocioInput{ClienteId: "cliente-1", Nome: "Negocio"}, "usuario-1", "usuario")

	if err == nil {
		t.Fatal("erro esperado de acesso")
	}
	if repository.salvo != nil {
		t.Fatal("negocio sem acesso nao deveria ser salvo")
	}
}

func TestAtualizarNegocioDetectaConflitoDeVersao(t *testing.T) {
	repository := &negocioRepositoryFake{negocio: &domain.Negocio{Id: "negocio-1", ClienteId: "cliente-1", Version: 2}}
	useCase := NewUseCase(repository, &clienteAcessoRepositoryFake{})

	_, err := useCase.AtualizarNegocio(context.Background(), "negocio-1", dto.AtualizarNegocioInput{Nome: "Novo", Version: 1}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrVersaoConflito) {
		t.Fatalf("erro esperado de versao, recebido %v", err)
	}
	if repository.atualizado != nil {
		t.Fatal("negocio com conflito nao deveria ser atualizado")
	}
}
