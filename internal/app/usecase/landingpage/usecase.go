package landingpage

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"nous/internal/app/dto"
	"nous/internal/domain/cliente"
	"nous/internal/domain/evento"
	domainexportacao "nous/internal/domain/exportacao"
	"nous/internal/domain/hosting"
	domain "nous/internal/domain/landingpage"
	"nous/internal/domain/plano"
	"nous/internal/domain/template"

	"github.com/google/uuid"
)

var (
	regexpNaoAlfa   = regexp.MustCompile(`[^a-z0-9]+`)
	regexpHifenMult = regexp.MustCompile(`-{2,}`)
)

type UseCase struct {
	repository           domain.Repository
	clienteRepository    cliente.Repository
	publisher            evento.Publisher
	planoRepository      plano.Repository
	templateRepository   template.Repository
	deployer             hosting.Deployer
	exportacaoRepository domainexportacao.Repository
}

func NewUseCase(
	repository domain.Repository,
	clienteRepository cliente.Repository,
	publisher evento.Publisher,
	planoRepository plano.Repository,
	templateRepository template.Repository,
	deployer hosting.Deployer,
	exportacaoRepository domainexportacao.Repository,
) *UseCase {
	return &UseCase{
		repository:           repository,
		clienteRepository:    clienteRepository,
		publisher:            publisher,
		planoRepository:      planoRepository,
		templateRepository:   templateRepository,
		deployer:             deployer,
		exportacaoRepository: exportacaoRepository,
	}
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

	slug := gerarSlug(input.Slug)
	if slug == "" {
		slug = gerarSlug(input.Nome)
	}
	if slug == "" {
		slug = uuid.New().String()[:8]
	}

	lp := &domain.LandingPage{
		ClienteId:    input.ClienteId,
		NegocioId:    input.NegocioId,
		CategoriaId:  categoriaId,
		TemplateId:   templateId,
		Tipo:         tipo,
		Nome:         input.Nome,
		Slug:         slug,
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
	if input.Slug != "" {
		lp.Slug = gerarSlug(input.Slug)
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

func (u *UseCase) PublicarLandingPage(ctx context.Context, id, usuarioId, role string) (*dto.PublicarLandingPageOutput, error) {
	if err := u.verificarAcessoLandingPage(ctx, id, usuarioId, role); err != nil {
		return nil, err
	}
	lp, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("publicar landingpage: %w", err)
	}
	if lp.Slug == "" {
		return nil, domain.ErrSlugNaoDefinido
	}

	exp, err := u.exportacaoRepository.BuscarMaisRecentePorLandingPage(ctx, id)
	if errors.Is(err, domainexportacao.ErrNaoEncontrada) {
		return nil, domain.ErrSemExportacao
	}
	if err != nil {
		return nil, fmt.Errorf("publicar landingpage buscar exportacao: %w", err)
	}

	arquivos, err := extrairArquivosZip(exp.Arquivo)
	if err != nil {
		return nil, fmt.Errorf("publicar landingpage extrair zip: %w", err)
	}

	resultado, err := u.deployer.Deploy(ctx, lp.Slug, arquivos)
	if err != nil {
		return nil, fmt.Errorf("publicar landingpage deploy: %w", err)
	}

	agora := time.Now()
	if err := u.repository.AtualizarPublicacao(ctx, id, resultado.UrlPublica, resultado.DeployId, agora); err != nil {
		return nil, fmt.Errorf("publicar landingpage salvar: %w", err)
	}

	_ = u.publisher.Publicar(ctx, evento.Evento{
		EventType:      "landingpage.publicada",
		Source:         "nous",
		SchemaVersion:  1,
		IdempotencyKey: "nous:landingpage:publicada:" + id,
		Payload: map[string]any{
			"lp_id":       id,
			"url_publica": resultado.UrlPublica,
			"deploy_id":   resultado.DeployId,
		},
	})

	return &dto.PublicarLandingPageOutput{
		Mensagem:   "landingpage publicada com sucesso",
		UrlPublica: resultado.UrlPublica,
		DeployId:   resultado.DeployId,
	}, nil
}

func (u *UseCase) DespublicarLandingPage(ctx context.Context, id, usuarioId, role string) error {
	if err := u.verificarAcessoLandingPage(ctx, id, usuarioId, role); err != nil {
		return err
	}
	lp, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return fmt.Errorf("despublicar landingpage: %w", err)
	}
	if lp.Status != domain.StatusPublicada {
		return domain.ErrNaoPublicada
	}

	if err := u.deployer.Despublicar(ctx, lp.Slug); err != nil {
		return fmt.Errorf("despublicar landingpage hosting: %w", err)
	}

	if err := u.repository.AtualizarDespublicacao(ctx, id); err != nil {
		return fmt.Errorf("despublicar landingpage salvar: %w", err)
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

func gerarSlug(texto string) string {
	texto = strings.TrimSpace(texto)
	if texto == "" {
		return ""
	}
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	normalizado, _, _ := transform.String(t, strings.ToLower(texto))
	slug := regexpNaoAlfa.ReplaceAllString(normalizado, "-")
	slug = regexpHifenMult.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func extrairArquivosZip(dados []byte) ([]hosting.Arquivo, error) {
	r, err := zip.NewReader(bytes.NewReader(dados), int64(len(dados)))
	if err != nil {
		return nil, fmt.Errorf("abrir zip: %w", err)
	}
	var arquivos []hosting.Arquivo
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		caminho := filepath.ToSlash(f.Name)
		if strings.Contains(caminho, "..") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("abrir arquivo %s: %w", caminho, err)
		}
		conteudo, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("ler arquivo %s: %w", caminho, err)
		}
		arquivos = append(arquivos, hosting.Arquivo{Caminho: caminho, Conteudo: conteudo})
	}
	return arquivos, nil
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
		Slug:           lp.Slug,
		Conteudo:       lp.Conteudo,
		Status:         string(lp.Status),
		IsDraft:        lp.IsDraft,
		AutosaveData:   lp.AutosaveData,
		LastAutosaveEm: lp.LastAutosaveEm,
		LockedBy:       lp.LockedBy,
		LockedAt:       lp.LockedAt,
		UrlPublica:     lp.UrlPublica,
		DeployId:       lp.DeployId,
		PublicadaEm:    lp.PublicadaEm,
		CriadoEm:       lp.CriadoEm,
		AtualizadoEm:   lp.AtualizadoEm,
		Version:        lp.Version,
	}
}
