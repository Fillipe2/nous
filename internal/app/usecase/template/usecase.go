package template

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"nous/internal/app/dto"
	"nous/internal/domain/negocio"
	"nous/internal/domain/plano"
	domain "nous/internal/domain/template"
)

type UseCase struct {
	repository        domain.Repository
	negocioRepository negocio.Repository
	planoRepository   plano.Repository
}

func NewUseCase(repository domain.Repository, negocioRepository negocio.Repository, planoRepository plano.Repository) *UseCase {
	return &UseCase{repository: repository, negocioRepository: negocioRepository, planoRepository: planoRepository}
}

func (u *UseCase) CriarTemplate(ctx context.Context, input dto.CriarTemplateInput, usuarioId, role string) (*dto.TemplateOutput, error) {
	if strings.TrimSpace(input.Nome) == "" {
		return nil, fmt.Errorf("nome e obrigatorio")
	}
	if strings.TrimSpace(input.Conteudo) == "" && len(input.Arquivo) == 0 {
		return nil, fmt.Errorf("informe o conteudo ou o arquivo do template")
	}

	origem, err := domain.NormalizarOrigem(input.Origem)
	if err != nil {
		return nil, err
	}
	if input.Origem == "" && input.NegocioId != "" {
		origem = domain.OrigemNegocio
	}
	tipo := domain.NormalizarTipo(input.Tipo)
	negocioId := strings.TrimSpace(input.NegocioId)

	if origem == domain.OrigemSognix {
		if role != "admin" {
			return nil, domain.ErrNaoAutorizado
		}
		negocioId = ""
	}
	if origem == domain.OrigemNegocio {
		if negocioId == "" {
			return nil, domain.ErrNegocioObrigatorio
		}
		if err := u.verificarAcessoNegocio(ctx, negocioId, usuarioId, role); err != nil {
			return nil, err
		}
		if tipo == domain.TipoBaseCustomizavel && !u.negocioPermiteCustomizavel(ctx, negocioId) {
			return nil, domain.ErrPlanoNaoPermite
		}
	}

	t := &domain.Template{
		NegocioId:    negocioId,
		CategoriaId:  input.CategoriaId,
		Tipo:         tipo,
		Origem:       origem,
		Nome:         input.Nome,
		ArquivoURL:   input.ArquivoURL,
		Conteudo:     input.Conteudo,
		Arquivo:      input.Arquivo,
		NomeArquivo:  input.NomeArquivo,
		ContentType:  input.ContentType,
		TamanhoBytes: input.TamanhoBytes,
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

// BaixarTemplate retorna os bytes do pacote (ZIP) de um template e seus metadados.
// Templates oficiais Sognix ficam disponíveis para qualquer usuário autenticado;
// templates de negocio exigem que o usuário tenha acesso ao negocio dono.
func (u *UseCase) BaixarTemplate(ctx context.Context, id, usuarioId, role string) ([]byte, *dto.TemplateOutput, error) {
	t, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("baixar template: %w", err)
	}
	if t.Origem == domain.OrigemNegocio && t.NegocioId != "" {
		if err := u.verificarAcessoNegocio(ctx, t.NegocioId, usuarioId, role); err != nil {
			return nil, nil, err
		}
	}
	if !t.TemArquivo() {
		return nil, nil, domain.ErrSemArquivo
	}
	return t.Arquivo, mapearTemplateOutput(t), nil
}

func (u *UseCase) ExcluirTemplate(ctx context.Context, id string) error {
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir template: %w", err)
	}
	return nil
}

func (u *UseCase) ListarTemplates(ctx context.Context, filtro domain.FiltroListagem, usuarioId, role string) (*dto.ListaTemplatesOutput, error) {
	if role != "admin" {
		if filtro.NegocioId == "" {
			return nil, domain.ErrNegocioObrigatorio
		}
		if err := u.verificarAcessoNegocio(ctx, filtro.NegocioId, usuarioId, role); err != nil {
			return nil, err
		}
	}

	filtro.PermitirBaseCustomizavel = filtro.NegocioId == "" || u.negocioPermiteCustomizavel(ctx, filtro.NegocioId)
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

func (u *UseCase) verificarAcessoNegocio(ctx context.Context, negocioId, usuarioId, role string) error {
	if role == "admin" {
		return nil
	}
	if u.negocioRepository == nil {
		return domain.ErrNaoAutorizado
	}
	if err := u.negocioRepository.VerificarAcessoUsuario(ctx, negocioId, usuarioId); err != nil {
		return fmt.Errorf("acesso negocio: %w", err)
	}
	return nil
}

func (u *UseCase) negocioPermiteCustomizavel(ctx context.Context, negocioId string) bool {
	if negocioId == "" || u.planoRepository == nil {
		return false
	}
	planoAtivo, err := u.planoRepository.BuscarPlanoAtivoPorNegocio(ctx, negocioId)
	if errors.Is(err, plano.ErrAssinaturaAtivaNaoEncontrada) {
		return false
	}
	return err == nil && planoAtivo.Limites.PermiteCustomizavel()
}

func mapearTemplateOutput(t *domain.Template) *dto.TemplateOutput {
	return &dto.TemplateOutput{
		Id:           t.Id,
		NegocioId:    t.NegocioId,
		CategoriaId:  t.CategoriaId,
		Tipo:         string(t.Tipo),
		Origem:       string(t.Origem),
		Nome:         t.Nome,
		ArquivoURL:   t.ArquivoURL,
		Conteudo:     t.Conteudo,
		TemArquivo:   t.TemArquivo(),
		NomeArquivo:  t.NomeArquivo,
		ContentType:  t.ContentType,
		TamanhoBytes: t.TamanhoBytes,
		Status:       string(t.Status),
		CriadoEm:     t.CriadoEm,
		AtualizadoEm: t.AtualizadoEm,
	}
}
