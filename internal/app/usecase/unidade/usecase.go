package unidade

import (
	"context"
	"fmt"
	"time"

	"nous/internal/app/dto"
	"nous/internal/domain/negocio"
	domain "nous/internal/domain/unidade"
)

type UseCase struct {
	repository        domain.Repository
	negocioRepository negocio.Repository
}

func NewUseCase(repository domain.Repository, negocioRepository negocio.Repository) *UseCase {
	return &UseCase{repository: repository, negocioRepository: negocioRepository}
}

func (u *UseCase) CriarUnidade(ctx context.Context, input dto.CriarUnidadeInput, usuarioId, role string) (*dto.UnidadeOutput, error) {
	if input.NegocioId == "" || input.Nome == "" {
		return nil, fmt.Errorf("negocio_id e nome sao obrigatorios")
	}
	if err := u.verificarAcessoNegocio(ctx, input.NegocioId, usuarioId, role); err != nil {
		return nil, err
	}

	n := &domain.Unidade{
		NegocioId:    input.NegocioId,
		Nome:         input.Nome,
		Status:       domain.StatusAtivo,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
		Version:      1,
	}

	if err := u.repository.Salvar(ctx, n); err != nil {
		return nil, fmt.Errorf("criar unidade: %w", err)
	}

	return mapearUnidadeOutput(n), nil
}

func (u *UseCase) ObterUnidade(ctx context.Context, id string) (*dto.UnidadeOutput, error) {
	n, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter unidade: %w", err)
	}
	return mapearUnidadeOutput(n), nil
}

func (u *UseCase) AtualizarUnidade(ctx context.Context, id string, input dto.AtualizarUnidadeInput, usuarioId, role string) (*dto.UnidadeOutput, error) {
	n, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("atualizar unidade: %w", err)
	}
	if err := u.verificarAcessoNegocio(ctx, n.NegocioId, usuarioId, role); err != nil {
		return nil, err
	}
	if input.Version != n.Version {
		return nil, domain.ErrVersaoConflito
	}

	if input.Nome != "" {
		n.Nome = input.Nome
	}
	n.AtualizadoEm = time.Now()
	n.Version++

	if err := u.repository.Atualizar(ctx, n); err != nil {
		return nil, fmt.Errorf("atualizar unidade salvar: %w", err)
	}

	return mapearUnidadeOutput(n), nil
}

func (u *UseCase) ExcluirUnidade(ctx context.Context, id, usuarioId, role string) error {
	n, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return fmt.Errorf("excluir unidade: %w", err)
	}
	if err := u.verificarAcessoNegocio(ctx, n.NegocioId, usuarioId, role); err != nil {
		return err
	}
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir unidade: %w", err)
	}
	return nil
}

func (u *UseCase) ListarUnidades(ctx context.Context, filtro domain.FiltroListagem, usuarioId, role string) (*dto.ListaUnidadesOutput, error) {
	if role != "admin" {
		filtro.UsuarioId = usuarioId
		if filtro.NegocioId != "" {
			if err := u.verificarAcessoNegocio(ctx, filtro.NegocioId, usuarioId, role); err != nil {
				return nil, err
			}
		}
	}
	unidades, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar unidades: %w", err)
	}

	var saida []*dto.UnidadeOutput
	for _, n := range unidades {
		saida = append(saida, mapearUnidadeOutput(n))
	}

	return &dto.ListaUnidadesOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func (u *UseCase) verificarAcessoNegocio(ctx context.Context, negocioId, usuarioId, role string) error {
	if role == "admin" {
		return nil
	}
	if err := u.negocioRepository.VerificarAcessoUsuario(ctx, negocioId, usuarioId); err != nil {
		return fmt.Errorf("acesso negocio: %w", err)
	}
	return nil
}

func mapearUnidadeOutput(n *domain.Unidade) *dto.UnidadeOutput {
	return &dto.UnidadeOutput{
		Id:           n.Id,
		NegocioId:    n.NegocioId,
		Nome:         n.Nome,
		Status:       string(n.Status),
		CriadoEm:     n.CriadoEm,
		AtualizadoEm: n.AtualizadoEm,
		Version:      n.Version,
	}
}
