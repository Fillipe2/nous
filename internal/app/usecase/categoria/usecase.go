package categoria

import (
	"context"
	"fmt"
	"time"

	"nous/internal/app/dto"
	domain "nous/internal/domain/categoria"
)

type UseCase struct {
	repository domain.Repository
}

func NewUseCase(repository domain.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) CriarCategoria(ctx context.Context, input dto.CriarCategoriaInput) (*dto.CategoriaOutput, error) {
	if input.Nome == "" {
		return nil, fmt.Errorf("nome e obrigatorio")
	}

	c := &domain.Categoria{
		Nome:         input.Nome,
		Status:       domain.StatusAtivo,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
		Version:      1,
	}

	if err := u.repository.Salvar(ctx, c); err != nil {
		return nil, fmt.Errorf("criar categoria: %w", err)
	}

	return mapearCategoriaOutput(c), nil
}

func (u *UseCase) ObterCategoria(ctx context.Context, id string) (*dto.CategoriaOutput, error) {
	c, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter categoria: %w", err)
	}
	return mapearCategoriaOutput(c), nil
}

func (u *UseCase) AtualizarCategoria(ctx context.Context, id string, input dto.AtualizarCategoriaInput) (*dto.CategoriaOutput, error) {
	c, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("atualizar categoria: %w", err)
	}
	if input.Version != c.Version {
		return nil, domain.ErrVersaoConflito
	}

	if input.Nome != "" {
		c.Nome = input.Nome
	}
	c.AtualizadoEm = time.Now()
	c.Version++

	if err := u.repository.Atualizar(ctx, c); err != nil {
		return nil, fmt.Errorf("atualizar categoria salvar: %w", err)
	}

	return mapearCategoriaOutput(c), nil
}

func (u *UseCase) ExcluirCategoria(ctx context.Context, id string) error {
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir categoria: %w", err)
	}
	return nil
}

func (u *UseCase) ListarCategorias(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaCategoriasOutput, error) {
	categorias, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar categorias: %w", err)
	}

	var saida []*dto.CategoriaOutput
	for _, c := range categorias {
		saida = append(saida, mapearCategoriaOutput(c))
	}

	return &dto.ListaCategoriasOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func mapearCategoriaOutput(c *domain.Categoria) *dto.CategoriaOutput {
	return &dto.CategoriaOutput{
		Id:           c.Id,
		Nome:         c.Nome,
		Status:       string(c.Status),
		CriadoEm:     c.CriadoEm,
		AtualizadoEm: c.AtualizadoEm,
		Version:      c.Version,
	}
}
