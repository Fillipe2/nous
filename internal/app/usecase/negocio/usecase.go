package negocio

import (
	"context"
	"fmt"
	"time"

	"nous/internal/app/dto"
	"nous/internal/domain/cliente"
	domain "nous/internal/domain/negocio"
)

type UseCase struct {
	repository        domain.Repository
	clienteRepository cliente.Repository
}

func NewUseCase(repository domain.Repository, clienteRepository cliente.Repository) *UseCase {
	return &UseCase{repository: repository, clienteRepository: clienteRepository}
}

func (u *UseCase) CriarNegocio(ctx context.Context, input dto.CriarNegocioInput, usuarioId, role string) (*dto.NegocioOutput, error) {
	if input.ClienteId == "" || input.Nome == "" {
		return nil, fmt.Errorf("cliente_id e nome sao obrigatorios")
	}
	if err := u.verificarAcessoCliente(ctx, input.ClienteId, usuarioId, role); err != nil {
		return nil, err
	}

	n := &domain.Negocio{
		ClienteId:    input.ClienteId,
		Nome:         input.Nome,
		Status:       domain.StatusAtivo,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
		Version:      1,
	}

	if err := u.repository.Salvar(ctx, n); err != nil {
		return nil, fmt.Errorf("criar negocio: %w", err)
	}

	return mapearNegocioOutput(n), nil
}

func (u *UseCase) ObterNegocio(ctx context.Context, id string) (*dto.NegocioOutput, error) {
	n, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter negocio: %w", err)
	}
	return mapearNegocioOutput(n), nil
}

func (u *UseCase) AtualizarNegocio(ctx context.Context, id string, input dto.AtualizarNegocioInput, usuarioId, role string) (*dto.NegocioOutput, error) {
	n, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("atualizar negocio: %w", err)
	}
	if err := u.verificarAcessoCliente(ctx, n.ClienteId, usuarioId, role); err != nil {
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
		return nil, fmt.Errorf("atualizar negocio salvar: %w", err)
	}

	return mapearNegocioOutput(n), nil
}

func (u *UseCase) ExcluirNegocio(ctx context.Context, id, usuarioId, role string) error {
	n, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return fmt.Errorf("excluir negocio: %w", err)
	}
	if err := u.verificarAcessoCliente(ctx, n.ClienteId, usuarioId, role); err != nil {
		return err
	}
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir negocio: %w", err)
	}
	return nil
}

func (u *UseCase) ListarNegocios(ctx context.Context, filtro domain.FiltroListagem, usuarioId, role string) (*dto.ListaNegociosOutput, error) {
	if role != "admin" {
		filtro.UsuarioId = usuarioId
		if filtro.ClienteId != "" {
			if err := u.verificarAcessoCliente(ctx, filtro.ClienteId, usuarioId, role); err != nil {
				return nil, err
			}
		}
	}
	negocios, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar negocios: %w", err)
	}

	var saida []*dto.NegocioOutput
	for _, n := range negocios {
		saida = append(saida, mapearNegocioOutput(n))
	}

	return &dto.ListaNegociosOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func (u *UseCase) verificarAcessoCliente(ctx context.Context, clienteId, usuarioId, role string) error {
	if role == "admin" {
		return nil
	}
	if err := u.clienteRepository.VerificarAcessoUsuario(ctx, clienteId, usuarioId); err != nil {
		return fmt.Errorf("acesso cliente: %w", err)
	}
	return nil
}

func mapearNegocioOutput(n *domain.Negocio) *dto.NegocioOutput {
	return &dto.NegocioOutput{
		Id:           n.Id,
		ClienteId:    n.ClienteId,
		Nome:         n.Nome,
		Status:       string(n.Status),
		CriadoEm:     n.CriadoEm,
		AtualizadoEm: n.AtualizadoEm,
		Version:      n.Version,
	}
}
