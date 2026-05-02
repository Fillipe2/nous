package landingpage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/evento"
	domain "buildnix-backend/internal/domain/landingpage"
	"buildnix-backend/internal/domain/plano"
	"buildnix-backend/internal/domain/template"
)

type UseCase struct {
	repository         domain.Repository
	clienteRepository  cliente.Repository
	publisher          evento.Publisher
	planoRepository    plano.Repository
	templateRepository template.Repository
}

func NewUseCase(repository domain.Repository, clienteRepository cliente.Repository, publisher evento.Publisher, planoRepository plano.Repository, templateRepository template.Repository) *UseCase {
	return &UseCase{repository: repository, clienteRepository: clienteRepository, publisher: publisher, planoRepository: planoRepository, templateRepository: templateRepository}
}

func (u *UseCase) CriarLandingPage(ctx context.Context, input dto.CriarLandingPageInput, usuarioId, role string) (*dto.LandingPageOutput, error) {
	if input.ClienteId == "" || input.Nome == "" {
		return nil, fmt.Errorf("cliente_id e nome sao obrigatorios")
	}
	if err := u.verificarAcessoCliente(ctx, input.ClienteId, usuarioId, role); err != nil {
		return nil, err
	}
	tipo, err := domain.NormalizarTipo(input.Tipo)
	if err != nil {
		return nil, err
	}
	if err := u.validarPlanoLandingPage(ctx, input.ClienteId, tipo); err != nil {
		return nil, err
	}
	if err := u.validarLimiteLandingPages(ctx, input.ClienteId); err != nil {
		return nil, err
	}
	conteudo, templateId, categoriaId, err := u.prepararTemplateInicial(ctx, input.ClienteId, tipo, input.TemplateId, input.CategoriaId, input.Conteudo)
	if err != nil {
		return nil, err
	}

	lp := &domain.LandingPage{
		ClienteId:    input.ClienteId,
		NegocioId:    input.NegocioId,
		CategoriaId:  categoriaId,
		TemplateId:   templateId,
		Tipo:         tipo,
		Nome:         input.Nome,
		Conteudo:     conteudo,
		Status:       domain.StatusRascunho,
		IsDraft:      true,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
		Version:      1,
	}

	if err := u.repository.Salvar(ctx, lp); err != nil {
		return nil, fmt.Errorf("criar landingpage: %w", err)
	}

	return mapearLandingPageOutput(lp), nil
}

func (u *UseCase) ObterLandingPage(ctx context.Context, id, usuarioId, role string) (*dto.LandingPageOutput, error) {
	lp, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter landingpage: %w", err)
	}
	if err := u.verificarAcessoLandingPage(ctx, id, usuarioId, role); err != nil {
		return nil, err
	}
	return mapearLandingPageOutput(lp), nil
}

func (u *UseCase) AtualizarLandingPage(ctx context.Context, id string, input dto.AtualizarLandingPageInput, usuarioId, role string) (*dto.LandingPageOutput, error) {
	lp, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("atualizar landingpage: %w", err)
	}
	if err := u.verificarAcessoLandingPage(ctx, id, usuarioId, role); err != nil {
		return nil, err
	}
	tipo := lp.Tipo
	if input.Tipo != "" {
		tipo, err = domain.NormalizarTipo(input.Tipo)
		if err != nil {
			return nil, err
		}
		if err := u.validarPlanoLandingPage(ctx, lp.ClienteId, tipo); err != nil {
			return nil, err
		}
	}

	if lp.Version != input.Version {
		return nil, domain.ErrVersaoConflito
	}
	if tipo == domain.TipoWhiteLabel && input.Conteudo != "" {
		return nil, domain.ErrPlanoNaoPermite
	}

	if input.Nome != "" {
		lp.Nome = input.Nome
	}
	if input.Conteudo != "" {
		lp.Conteudo = input.Conteudo
	}
	if input.NegocioId != "" {
		lp.NegocioId = input.NegocioId
	}
	if input.CategoriaId != "" {
		lp.CategoriaId = input.CategoriaId
	}
	if input.TemplateId != "" {
		conteudo, templateId, categoriaId, err := u.prepararTemplateInicial(ctx, lp.ClienteId, tipo, input.TemplateId, lp.CategoriaId, lp.Conteudo)
		if err != nil {
			return nil, err
		}
		lp.TemplateId = templateId
		lp.CategoriaId = categoriaId
		if tipo == domain.TipoWhiteLabel || input.Conteudo == "" {
			lp.Conteudo = conteudo
		}
	}
	lp.Tipo = tipo
	if input.Status != "" {
		lp.Status = domain.Status(input.Status)
		if lp.Status == domain.StatusPublicada {
			lp.IsDraft = false
		}
	}
	lp.AtualizadoEm = time.Now()
	lp.Version++

	if err := u.repository.Atualizar(ctx, lp); err != nil {
		return nil, fmt.Errorf("atualizar landingpage salvar: %w", err)
	}

	return mapearLandingPageOutput(lp), nil
}

func (u *UseCase) ExcluirLandingPage(ctx context.Context, id, usuarioId, role string) error {
	if err := u.verificarAcessoLandingPage(ctx, id, usuarioId, role); err != nil {
		return err
	}
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir landingpage: %w", err)
	}
	return nil
}

func (u *UseCase) ListarLandingPages(ctx context.Context, filtro domain.FiltroListagem, usuarioId, role string) (*dto.ListaLandingPagesOutput, error) {
	if role != "admin" {
		filtro.UsuarioId = usuarioId
		if filtro.ClienteId != "" {
			if err := u.clienteRepository.VerificarAcessoUsuario(ctx, filtro.ClienteId, usuarioId); err != nil {
				return nil, fmt.Errorf("listar landingpages acesso cliente: %w", err)
			}
		}
	}
	landingPages, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar landingpages: %w", err)
	}

	var saida []*dto.LandingPageOutput
	for _, lp := range landingPages {
		saida = append(saida, mapearLandingPageOutput(lp))
	}

	return &dto.ListaLandingPagesOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func (u *UseCase) Autosave(ctx context.Context, id, usuarioId, role string, input dto.AutosaveInput) (*dto.AutosaveOutput, error) {
	if err := u.verificarAcessoLandingPage(ctx, id, usuarioId, role); err != nil {
		return nil, fmt.Errorf("autosave acesso negado: %w", err)
	}
	lp, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("autosave obter landingpage: %w", err)
	}
	if lp.Tipo != domain.TipoCustomizavel {
		return nil, domain.ErrPlanoNaoPermite
	}
	if err := u.validarPlanoLandingPage(ctx, lp.ClienteId, domain.TipoCustomizavel); err != nil {
		return nil, err
	}
	if len(input.AutosaveData) == 0 {
		return nil, fmt.Errorf("autosave_data nao pode ser vazio")
	}
	if input.Version <= 0 {
		return nil, fmt.Errorf("version e obrigatorio")
	}
	novaVersao, err := u.repository.AtualizarAutosave(ctx, id, input.AutosaveData, input.Version)
	if err != nil {
		return nil, fmt.Errorf("autosave landingpage: %w", err)
	}
	return &dto.AutosaveOutput{Mensagem: "rascunho salvo", Version: novaVersao}, nil
}

func (u *UseCase) BloquearEdicao(ctx context.Context, id, usuarioId, role string) (*dto.LandingPageOutput, error) {
	if err := u.verificarAcessoLandingPage(ctx, id, usuarioId, role); err != nil {
		return nil, fmt.Errorf("bloquear edicao acesso negado: %w", err)
	}
	lp, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bloquear edicao: %w", err)
	}

	agora := time.Now()
	if lp.LockedBy != nil && *lp.LockedBy != usuarioId {
		if lp.LockedAt != nil && agora.Before(lp.LockedAt.Add(domain.DuracaoLock)) {
			return nil, domain.ErrBloqueada
		}
	}

	lp.LockedBy = &usuarioId
	lp.LockedAt = &agora

	if err := u.repository.AtualizarLock(ctx, id, &usuarioId); err != nil {
		return nil, fmt.Errorf("bloquear edicao salvar: %w", err)
	}

	return mapearLandingPageOutput(lp), nil
}

func (u *UseCase) DesbloquearEdicao(ctx context.Context, id, usuarioId, role string) error {
	if err := u.verificarAcessoLandingPage(ctx, id, usuarioId, role); err != nil {
		return fmt.Errorf("desbloquear edicao acesso negado: %w", err)
	}
	lp, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return fmt.Errorf("desbloquear edicao: %w", err)
	}

	if lp.LockedBy != nil && *lp.LockedBy != usuarioId {
		return domain.ErrNaoEncontrada
	}

	if err := u.repository.AtualizarLock(ctx, id, nil); err != nil {
		return fmt.Errorf("desbloquear edicao salvar: %w", err)
	}

	return nil
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

func (u *UseCase) verificarAcessoLandingPage(ctx context.Context, id, usuarioId, role string) error {
	if role == "admin" {
		return nil
	}
	if err := u.repository.VerificarAcessoUsuario(ctx, id, usuarioId); err != nil {
		return fmt.Errorf("acesso landingpage: %w", err)
	}
	return nil
}

func (u *UseCase) validarPlanoLandingPage(ctx context.Context, clienteId string, tipo domain.Tipo) error {
	if tipo == domain.TipoWhiteLabel && u.planoRepository == nil {
		return nil
	}
	if u.planoRepository == nil {
		return domain.ErrPlanoNaoPermite
	}
	planoAtivo, err := u.planoRepository.BuscarPlanoAtivoPorCliente(ctx, clienteId)
	if errors.Is(err, plano.ErrAssinaturaAtivaNaoEncontrada) {
		if tipo == domain.TipoWhiteLabel {
			return nil
		}
		return domain.ErrPlanoNaoPermite
	}
	if err != nil {
		return fmt.Errorf("validar plano landingpage: %w", err)
	}
	if tipo == domain.TipoWhiteLabel && !planoAtivo.Limites.PermiteWhiteLabel() {
		return domain.ErrPlanoNaoPermite
	}
	if tipo == domain.TipoCustomizavel && !planoAtivo.Limites.PermiteCustomizavel() {
		return domain.ErrPlanoNaoPermite
	}
	return nil
}

func (u *UseCase) validarLimiteLandingPages(ctx context.Context, clienteId string) error {
	if u.planoRepository == nil {
		return nil
	}
	planoAtivo, err := u.planoRepository.BuscarPlanoAtivoPorCliente(ctx, clienteId)
	if errors.Is(err, plano.ErrAssinaturaAtivaNaoEncontrada) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("validar limite landingpages: %w", err)
	}
	if planoAtivo.Limites.MaxLandingPages <= 0 {
		return nil
	}
	total, err := u.repository.ContarPorCliente(ctx, clienteId)
	if err != nil {
		return fmt.Errorf("contar landingpages do cliente: %w", err)
	}
	if total >= planoAtivo.Limites.MaxLandingPages {
		return domain.ErrPlanoNaoPermite
	}
	return nil
}

func (u *UseCase) prepararTemplateInicial(ctx context.Context, clienteId string, tipo domain.Tipo, templateId, categoriaId, conteudo string) (string, string, string, error) {
	templateId = strings.TrimSpace(templateId)
	if tipo == domain.TipoWhiteLabel && templateId == "" {
		return "", "", "", domain.ErrTemplateObrigatorio
	}
	if templateId == "" {
		return conteudo, "", categoriaId, nil
	}
	if u.templateRepository == nil {
		return "", "", "", domain.ErrTemplateObrigatorio
	}
	t, err := u.templateRepository.BuscarPorID(ctx, templateId)
	if err != nil {
		return "", "", "", fmt.Errorf("buscar template inicial: %w", err)
	}
	if tipo == domain.TipoWhiteLabel && t.Tipo != template.TipoWhiteLabel {
		return "", "", "", domain.ErrTemplateIncompativel
	}
	if t.Origem == template.OrigemCliente && t.ClienteId != clienteId {
		return "", "", "", domain.ErrTemplateIncompativel
	}
	if strings.TrimSpace(conteudo) == "" {
		conteudo = t.Conteudo
	}
	if categoriaId == "" {
		categoriaId = t.CategoriaId
	}
	return conteudo, templateId, categoriaId, nil
}

func mapearLandingPageOutput(lp *domain.LandingPage) *dto.LandingPageOutput {
	return &dto.LandingPageOutput{
		Id:             lp.Id,
		ClienteId:      lp.ClienteId,
		NegocioId:      lp.NegocioId,
		CategoriaId:    lp.CategoriaId,
		TemplateId:     lp.TemplateId,
		Tipo:           string(lp.Tipo),
		Nome:           lp.Nome,
		Conteudo:       lp.Conteudo,
		Status:         string(lp.Status),
		IsDraft:        lp.IsDraft,
		AutosaveData:   lp.AutosaveData,
		LastAutosaveEm: lp.LastAutosaveEm,
		LockedBy:       lp.LockedBy,
		LockedAt:       lp.LockedAt,
		CriadoEm:       lp.CriadoEm,
		AtualizadoEm:   lp.AtualizadoEm,
		Version:        lp.Version,
	}
}
