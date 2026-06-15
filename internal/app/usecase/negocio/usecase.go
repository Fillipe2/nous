package negocio

import (
	"context"
	"fmt"
	"time"

	"nous/internal/app/dto"
	domain "nous/internal/domain/negocio"
)

type UseCase struct {
	repository domain.Repository
}

func NewUseCase(repository domain.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) CriarNegocio(ctx context.Context, input dto.CriarNegocioInput, usuarioId string) (*dto.NegocioOutput, error) {
	if input.Nome == "" {
		return nil, fmt.Errorf("nome e obrigatorio")
	}
	if usuarioId == "" {
		return nil, domain.ErrNaoAutorizado
	}

	documento := domain.NormalizarDocumento(input.Documento)
	if documento != "" {
		if err := domain.ValidarDocumento(documento); err != nil {
			return nil, err
		}
	}

	c := &domain.Negocio{
		Nome:         input.Nome,
		Email:        input.Email,
		Telefone:     input.Telefone,
		Whatsapp:     input.Whatsapp,
		Documento:    documento,
		Status:       domain.StatusAtivo,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
		Version:      1,
	}

	if err := u.repository.SalvarComVinculo(ctx, c, &domain.UsuarioNegocio{
		UsuarioId: usuarioId,
		CriadoEm:  time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("criar negocio: %w", err)
	}

	return mapearNegocioOutput(c), nil
}

func (u *UseCase) ObterNegocio(ctx context.Context, id string) (*dto.NegocioOutput, error) {
	c, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter negocio: %w", err)
	}
	return mapearNegocioOutput(c), nil
}

func (u *UseCase) AtualizarNegocio(ctx context.Context, id string, input dto.AtualizarNegocioInput) (*dto.NegocioOutput, error) {
	c, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("atualizar negocio: %w", err)
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
	if input.Documento != "" {
		documento := domain.NormalizarDocumento(input.Documento)
		if err := domain.ValidarDocumento(documento); err != nil {
			return nil, err
		}
		c.Documento = documento
	}
	c.AtualizadoEm = time.Now()
	c.Version++

	if err := u.repository.Atualizar(ctx, c); err != nil {
		return nil, fmt.Errorf("atualizar negocio salvar: %w", err)
	}

	return mapearNegocioOutput(c), nil
}

func (u *UseCase) ExcluirNegocio(ctx context.Context, id string) error {
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir negocio: %w", err)
	}
	return nil
}

func (u *UseCase) ListarNegocios(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaNegociosOutput, error) {
	negocios, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar negocios: %w", err)
	}

	var saida []*dto.NegocioOutput
	for _, c := range negocios {
		saida = append(saida, mapearNegocioOutput(c))
	}

	return &dto.ListaNegociosOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func (u *UseCase) ListarNegociosDoUsuario(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaNegociosBasicosOutput, error) {
	negocios, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar negocios do usuario: %w", err)
	}

	var saida []*dto.NegocioBasicoOutput
	for _, c := range negocios {
		saida = append(saida, mapearNegocioBasicoOutput(c))
	}

	return &dto.ListaNegociosBasicosOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func (u *UseCase) VincularUsuario(ctx context.Context, input dto.VincularUsuarioNegocioInput) error {
	uc := &domain.UsuarioNegocio{
		UsuarioId: input.UsuarioId,
		NegocioId: input.NegocioId,
		CriadoEm:  time.Now(),
	}
	if err := u.repository.VincularUsuario(ctx, uc); err != nil {
		return fmt.Errorf("vincular usuario negocio: %w", err)
	}
	return nil
}

func mapearNegocioOutput(c *domain.Negocio) *dto.NegocioOutput {
	return &dto.NegocioOutput{
		Id:           c.Id,
		Nome:         c.Nome,
		Email:        c.Email,
		Telefone:     c.Telefone,
		Whatsapp:     c.Whatsapp,
		Documento:    c.Documento,
		Status:       string(c.Status),
		CriadoEm:     c.CriadoEm,
		AtualizadoEm: c.AtualizadoEm,
		Version:      c.Version,
	}
}

func mapearNegocioBasicoOutput(c *domain.Negocio) *dto.NegocioBasicoOutput {
	return &dto.NegocioBasicoOutput{
		Id:        c.Id,
		Nome:      c.Nome,
		Email:     c.Email,
		Telefone:  c.Telefone,
		Whatsapp:  c.Whatsapp,
		Documento: c.Documento,
		Status:    string(c.Status),
	}
}
