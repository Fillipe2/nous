package template

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/plano"
	domain "buildnix-backend/internal/domain/template"
)

type UseCase struct {
	repository        domain.Repository
	clienteRepository cliente.Repository
	planoRepository   plano.Repository
}

func NewUseCase(repository domain.Repository, clienteRepository cliente.Repository, planoRepository plano.Repository) *UseCase {
	return &UseCase{repository: repository, clienteRepository: clienteRepository, planoRepository: planoRepository}
}

func (u *UseCase) CriarTemplate(ctx context.Context, input dto.CriarTemplateInput, usuarioId, role string) (*dto.TemplateOutput, error) {
	if strings.TrimSpace(input.Nome) == "" {
		return nil, fmt.Errorf("nome e obrigatorio")
	}
	if strings.TrimSpace(input.Conteudo) == "" {
		return nil, fmt.Errorf("conteudo do template e obrigatorio")
	}

	origem, err := domain.NormalizarOrigem(input.Origem)
	if err != nil {
		return nil, err
	}
	if input.Origem == "" && input.ClienteId != "" {
		origem = domain.OrigemCliente
	}
	tipo := domain.NormalizarTipo(input.Tipo)
	clienteId := strings.TrimSpace(input.ClienteId)

	if origem == domain.OrigemSognix {
		if role != "admin" {
			return nil, domain.ErrNaoAutorizado
		}
		clienteId = ""
	}
	if origem == domain.OrigemCliente {
		if clienteId == "" {
			return nil, domain.ErrClienteObrigatorio
		}
		if err := u.verificarAcessoCliente(ctx, clienteId, usuarioId, role); err != nil {
			return nil, err
		}
		if tipo == domain.TipoBaseCustomizavel && !u.clientePermiteCustomizavel(ctx, clienteId) {
			return nil, domain.ErrPlanoNaoPermite
		}
	}

	t := &domain.Template{
		ClienteId:    clienteId,
		CategoriaId:  input.CategoriaId,
		Tipo:         tipo,
		Origem:       origem,
		Nome:         input.Nome,
		ArquivoURL:   input.ArquivoURL,
		Conteudo:     input.Conteudo,
		Status:       domain.StatusAtivo,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
		Version:      1,
	}

	if err := u.repository.Salvar(ctx, t); err != nil {
		return nil, fmt.Errorf("criar template: %w", err)
	}

	return mapearTemplateOutput(t), nil
}

func (u *UseCase) ObterTemplate(ctx context.Context, id string) (*dto.TemplateOutput, error) {
	t, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter template: %w", err)
	}
	return mapearTemplateOutput(t), nil
}

func (u *UseCase) ExcluirTemplate(ctx context.Context, id string) error {
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir template: %w", err)
	}
	return nil
}

func (u *UseCase) ListarTemplates(ctx context.Context, filtro domain.FiltroListagem, usuarioId, role string) (*dto.ListaTemplatesOutput, error) {
	if role != "admin" {
		if filtro.ClienteId == "" {
			return nil, domain.ErrClienteObrigatorio
		}
		if err := u.verificarAcessoCliente(ctx, filtro.ClienteId, usuarioId, role); err != nil {
			return nil, err
		}
	}

	filtro.PermitirBaseCustomizavel = filtro.ClienteId == "" || u.clientePermiteCustomizavel(ctx, filtro.ClienteId)
	templates, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar templates: %w", err)
	}

	var saida []*dto.TemplateOutput
	for _, t := range templates {
		saida = append(saida, mapearTemplateOutput(t))
	}

	return &dto.ListaTemplatesOutput{
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
	if u.clienteRepository == nil {
		return domain.ErrNaoAutorizado
	}
	if err := u.clienteRepository.VerificarAcessoUsuario(ctx, clienteId, usuarioId); err != nil {
		return fmt.Errorf("acesso cliente: %w", err)
	}
	return nil
}

func (u *UseCase) clientePermiteCustomizavel(ctx context.Context, clienteId string) bool {
	if clienteId == "" || u.planoRepository == nil {
		return false
	}
	planoAtivo, err := u.planoRepository.BuscarPlanoAtivoPorCliente(ctx, clienteId)
	if errors.Is(err, plano.ErrAssinaturaAtivaNaoEncontrada) {
		return false
	}
	return err == nil && planoAtivo.Limites.PermiteCustomizavel()
}

func mapearTemplateOutput(t *domain.Template) *dto.TemplateOutput {
	return &dto.TemplateOutput{
		Id:           t.Id,
		ClienteId:    t.ClienteId,
		CategoriaId:  t.CategoriaId,
		Tipo:         string(t.Tipo),
		Origem:       string(t.Origem),
		Nome:         t.Nome,
		ArquivoURL:   t.ArquivoURL,
		Conteudo:     t.Conteudo,
		Status:       string(t.Status),
		CriadoEm:     t.CriadoEm,
		AtualizadoEm: t.AtualizadoEm,
	}
}
