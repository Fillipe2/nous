package cliente

import (
	"context"
	"fmt"
	"time"

	"buildnix-backend/internal/app/dto"
	domain "buildnix-backend/internal/domain/cliente"
)

type UseCase struct {
	repository domain.Repository
}

func NewUseCase(repository domain.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) CriarCliente(ctx context.Context, input dto.CriarClienteInput, usuarioId string) (*dto.ClienteOutput, error) {
	if input.Nome == "" {
		return nil, fmt.Errorf("nome e obrigatorio")
	}
	if usuarioId == "" {
		return nil, domain.ErrNaoAutorizado
	}

	c := &domain.Cliente{
		Nome:         input.Nome,
		Email:        input.Email,
		Telefone:     input.Telefone,
		Whatsapp:     input.Whatsapp,
		Status:       domain.StatusAtivo,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
		Version:      1,
	}

	if err := u.repository.Salvar(ctx, c); err != nil {
		return nil, fmt.Errorf("criar cliente: %w", err)
	}
	if err := u.repository.VincularUsuario(ctx, &domain.UsuarioCliente{
		UsuarioId: usuarioId,
		ClienteId: c.Id,
		CriadoEm:  time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("vincular usuario ao cliente: %w", err)
	}

	return mapearClienteOutput(c), nil
}

func (u *UseCase) ObterCliente(ctx context.Context, id string) (*dto.ClienteOutput, error) {
	c, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter cliente: %w", err)
	}
	return mapearClienteOutput(c), nil
}

func (u *UseCase) AtualizarCliente(ctx context.Context, id string, input dto.AtualizarClienteInput) (*dto.ClienteOutput, error) {
	c, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("atualizar cliente: %w", err)
	}
	if input.Version != c.Version {
		return nil, domain.ErrVersaoConflito
	}

	if input.Nome != "" {
		c.Nome = input.Nome
	}
	if input.Email != "" {
		c.Email = input.Email
	}
	if input.Telefone != "" {
		c.Telefone = input.Telefone
	}
	if input.Whatsapp != "" {
		c.Whatsapp = input.Whatsapp
	}
	c.AtualizadoEm = time.Now()
	c.Version++

	if err := u.repository.Atualizar(ctx, c); err != nil {
		return nil, fmt.Errorf("atualizar cliente salvar: %w", err)
	}

	return mapearClienteOutput(c), nil
}

func (u *UseCase) ExcluirCliente(ctx context.Context, id string) error {
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir cliente: %w", err)
	}
	return nil
}

func (u *UseCase) ListarClientes(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaClientesOutput, error) {
	clientes, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar clientes: %w", err)
	}

	var saida []*dto.ClienteOutput
	for _, c := range clientes {
		saida = append(saida, mapearClienteOutput(c))
	}

	return &dto.ListaClientesOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func (u *UseCase) ListarClientesDoUsuario(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaClientesBasicosOutput, error) {
	clientes, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar clientes do usuario: %w", err)
	}

	var saida []*dto.ClienteBasicoOutput
	for _, c := range clientes {
		saida = append(saida, mapearClienteBasicoOutput(c))
	}

	return &dto.ListaClientesBasicosOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func (u *UseCase) VincularUsuario(ctx context.Context, input dto.VincularUsuarioClienteInput) error {
	uc := &domain.UsuarioCliente{
		UsuarioId: input.UsuarioId,
		ClienteId: input.ClienteId,
		CriadoEm:  time.Now(),
	}
	if err := u.repository.VincularUsuario(ctx, uc); err != nil {
		return fmt.Errorf("vincular usuario cliente: %w", err)
	}
	return nil
}

func mapearClienteOutput(c *domain.Cliente) *dto.ClienteOutput {
	return &dto.ClienteOutput{
		Id:           c.Id,
		Nome:         c.Nome,
		Email:        c.Email,
		Telefone:     c.Telefone,
		Whatsapp:     c.Whatsapp,
		Status:       string(c.Status),
		CriadoEm:     c.CriadoEm,
		AtualizadoEm: c.AtualizadoEm,
		Version:      c.Version,
	}
}

func mapearClienteBasicoOutput(c *domain.Cliente) *dto.ClienteBasicoOutput {
	return &dto.ClienteBasicoOutput{
		Id:       c.Id,
		Nome:     c.Nome,
		Email:    c.Email,
		Telefone: c.Telefone,
		Whatsapp: c.Whatsapp,
		Status:   string(c.Status),
	}
}
