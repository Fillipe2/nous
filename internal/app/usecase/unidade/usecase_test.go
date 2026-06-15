package unidade

import (
	"context"
	"errors"
	"testing"

	"nous/internal/app/dto"
	"nous/internal/domain/negocio"
	domain "nous/internal/domain/unidade"
)

type unidadeRepositoryFake struct {
	unidade    *domain.Unidade
	salvo      *domain.Unidade
	atualizado *domain.Unidade
}

func (r *unidadeRepositoryFake) Salvar(_ context.Context, n *domain.Unidade) error {
	n.Id = "unidade-novo"
	r.salvo = n
	return nil
}

func (r *unidadeRepositoryFake) BuscarPorID(context.Context, string) (*domain.Unidade, error) {
	if r.unidade == nil {
		return nil, domain.ErrNaoEncontrado
	}
	return r.unidade, nil
}

func (r *unidadeRepositoryFake) Atualizar(_ context.Context, n *domain.Unidade) error {
	r.atualizado = n
	return nil
}

func (r *unidadeRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *unidadeRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.Unidade, int, error) {
	return nil, 0, nil
}

type negocioAcessoRepositoryFake struct{ err error }

func (r *negocioAcessoRepositoryFake) Salvar(context.Context, *negocio.Negocio) error { return nil }
func (r *negocioAcessoRepositoryFake) SalvarComVinculo(context.Context, *negocio.Negocio, *negocio.UsuarioNegocio) error {
	return nil
}
func (r *negocioAcessoRepositoryFake) BuscarPorID(context.Context, string) (*negocio.Negocio, error) {
	return nil, negocio.ErrNaoEncontrado
}
func (r *negocioAcessoRepositoryFake) Atualizar(context.Context, *negocio.Negocio) error { return nil }
func (r *negocioAcessoRepositoryFake) SoftDelete(context.Context, string) error          { return nil }
func (r *negocioAcessoRepositoryFake) Listar(context.Context, negocio.FiltroListagem) ([]*negocio.Negocio, int, error) {
	return nil, 0, nil
}
func (r *negocioAcessoRepositoryFake) VincularUsuario(context.Context, *negocio.UsuarioNegocio) error {
	return nil
}
func (r *negocioAcessoRepositoryFake) ListarPorUsuario(context.Context, string) ([]*negocio.Negocio, error) {
	return nil, nil
}
func (r *negocioAcessoRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return r.err
}

func TestCriarUnidadeValidaCamposObrigatorios(t *testing.T) {
	useCase := NewUseCase(&unidadeRepositoryFake{}, &negocioAcessoRepositoryFake{})

	_, err := useCase.CriarUnidade(context.Background(), dto.CriarUnidadeInput{NegocioId: "negocio-1"}, "usuario-1", "usuario")

	if err == nil {
		t.Fatal("erro esperado para nome vazio")
	}
}

func TestCriarUnidadeExigeAcessoAoNegocio(t *testing.T) {
	repository := &unidadeRepositoryFake{}
	useCase := NewUseCase(repository, &negocioAcessoRepositoryFake{err: negocio.ErrNaoAutorizado})

	_, err := useCase.CriarUnidade(context.Background(), dto.CriarUnidadeInput{NegocioId: "negocio-1", Nome: "Unidade"}, "usuario-1", "usuario")

	if err == nil {
		t.Fatal("erro esperado de acesso")
	}
	if repository.salvo != nil {
		t.Fatal("unidade sem acesso nao deveria ser salvo")
	}
}

func TestAtualizarUnidadeDetectaConflitoDeVersao(t *testing.T) {
	repository := &unidadeRepositoryFake{unidade: &domain.Unidade{Id: "unidade-1", NegocioId: "negocio-1", Version: 2}}
	useCase := NewUseCase(repository, &negocioAcessoRepositoryFake{})

	_, err := useCase.AtualizarUnidade(context.Background(), "unidade-1", dto.AtualizarUnidadeInput{Nome: "Novo", Version: 1}, "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrVersaoConflito) {
		t.Fatalf("erro esperado de versao, recebido %v", err)
	}
	if repository.atualizado != nil {
		t.Fatal("unidade com conflito nao deveria ser atualizado")
	}
}
