package categoria

import (
	"context"
	"errors"
	"testing"

	"nous/internal/app/dto"
	domain "nous/internal/domain/categoria"
)

type categoriaRepositoryFake struct {
	categoria  *domain.Categoria
	salva      *domain.Categoria
	atualizada *domain.Categoria
}

func (r *categoriaRepositoryFake) Salvar(_ context.Context, c *domain.Categoria) error {
	c.Id = "categoria-nova"
	r.salva = c
	return nil
}

func (r *categoriaRepositoryFake) BuscarPorID(context.Context, string) (*domain.Categoria, error) {
	if r.categoria == nil {
		return nil, domain.ErrNaoEncontrada
	}
	return r.categoria, nil
}

func (r *categoriaRepositoryFake) Atualizar(_ context.Context, c *domain.Categoria) error {
	r.atualizada = c
	return nil
}

func (r *categoriaRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *categoriaRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.Categoria, int, error) {
	return nil, 0, nil
}

func TestCriarCategoriaValidaNome(t *testing.T) {
	useCase := NewUseCase(&categoriaRepositoryFake{})

	_, err := useCase.CriarCategoria(context.Background(), dto.CriarCategoriaInput{})

	if err == nil {
		t.Fatal("erro esperado para nome vazio")
	}
}

func TestCriarCategoriaDefineAtivaEVersaoInicial(t *testing.T) {
	repository := &categoriaRepositoryFake{}
	useCase := NewUseCase(repository)

	output, err := useCase.CriarCategoria(context.Background(), dto.CriarCategoriaInput{Nome: "Categoria"})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Id != "categoria-nova" || output.Status != string(domain.StatusAtivo) || output.Version != 1 {
		t.Fatalf("saida inesperada: %+v", output)
	}
	if repository.salva == nil || repository.salva.Status != domain.StatusAtivo {
		t.Fatalf("categoria salva inesperada: %+v", repository.salva)
	}
}

func TestAtualizarCategoriaDetectaConflitoDeVersao(t *testing.T) {
	repository := &categoriaRepositoryFake{categoria: &domain.Categoria{Id: "categoria-1", Version: 2}}
	useCase := NewUseCase(repository)

	_, err := useCase.AtualizarCategoria(context.Background(), "categoria-1", dto.AtualizarCategoriaInput{Nome: "Nova", Version: 1})

	if !errors.Is(err, domain.ErrVersaoConflito) {
		t.Fatalf("erro esperado de versao, recebido %v", err)
	}
	if repository.atualizada != nil {
		t.Fatal("categoria com conflito nao deveria ser atualizada")
	}
}
