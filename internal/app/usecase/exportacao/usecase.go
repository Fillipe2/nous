package exportacao

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"nous/internal/app/dto"
	domain "nous/internal/domain/exportacao"
	"nous/internal/domain/landingpage"
)

type UseCase struct {
	repository            domain.Repository
	landingPageRepository landingpage.Repository
}

func NewUseCase(repository domain.Repository, landingPageRepository landingpage.Repository) *UseCase {
	return &UseCase{repository: repository, landingPageRepository: landingPageRepository}
}

func (u *UseCase) ExportarLandingPage(ctx context.Context, landingPageId, usuarioId, role, nomeArquivo, contentType string, dados []byte) (*dto.ExportacaoOutput, error) {
	if err := u.verificarAcessoLandingPage(ctx, landingPageId, usuarioId, role); err != nil {
		return nil, err
	}
	if len(dados) == 0 {
		return nil, fmt.Errorf("arquivo zip nao pode ser vazio")
	}
	if contentType == "" {
		contentType = "application/zip"
	}
	sum := sha256.Sum256(dados)

	e := &domain.Exportacao{
		LandingPageId: landingPageId,
		UsuarioId:     usuarioId,
		NomeArquivo:   nomeArquivo,
		ContentType:   contentType,
		TamanhoBytes:  int64(len(dados)),
		Arquivo:       dados,
		Hash:          fmt.Sprintf("%x", sum),
		CriadoEm:      time.Now(),
	}

	if err := u.repository.Salvar(ctx, e); err != nil {
		return nil, fmt.Errorf("exportar landingpage salvar registro: %w", err)
	}

	return mapearExportacaoOutput(e), nil
}

func (u *UseCase) ObterExportacao(ctx context.Context, id string) (*dto.ExportacaoOutput, error) {
	e, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter exportacao: %w", err)
	}
	return mapearExportacaoOutput(e), nil
}

func (u *UseCase) BaixarExportacao(ctx context.Context, id, usuarioId, role string) ([]byte, *dto.ExportacaoOutput, error) {
	e, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("baixar exportacao buscar: %w", err)
	}
	if err := u.verificarAcessoLandingPage(ctx, e.LandingPageId, usuarioId, role); err != nil {
		return nil, nil, err
	}

	return e.Arquivo, mapearExportacaoOutput(e), nil
}

func (u *UseCase) ListarExportacoes(ctx context.Context, filtro domain.FiltroListagem, usuarioId, role string) (*dto.ListaExportacoesOutput, error) {
	if role != "admin" && filtro.LandingPageId == "" {
		return nil, fmt.Errorf("landingpage_id e obrigatorio")
	}
	if filtro.Pagina <= 0 {
		filtro.Pagina = 1
	}
	if filtro.PorPagina <= 0 {
		filtro.PorPagina = 20
	}
	if filtro.LandingPageId != "" {
		if err := u.verificarAcessoLandingPage(ctx, filtro.LandingPageId, usuarioId, role); err != nil {
			return nil, err
		}
	}
	exportacoes, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar exportacoes: %w", err)
	}

	var saida []*dto.ExportacaoOutput
	for _, e := range exportacoes {
		saida = append(saida, mapearExportacaoOutput(e))
	}

	return &dto.ListaExportacoesOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func (u *UseCase) verificarAcessoLandingPage(ctx context.Context, landingPageId, usuarioId, role string) error {
	if role == "admin" {
		return nil
	}
	if err := u.landingPageRepository.VerificarAcessoUsuario(ctx, landingPageId, usuarioId); err != nil {
		if errors.Is(err, landingpage.ErrNaoAutorizado) {
			return domain.ErrNaoAutorizada
		}
		return fmt.Errorf("acesso landingpage: %w", err)
	}
	return nil
}

func mapearExportacaoOutput(e *domain.Exportacao) *dto.ExportacaoOutput {
	return &dto.ExportacaoOutput{
		Id:            e.Id,
		LandingPageId: e.LandingPageId,
		UsuarioId:     e.UsuarioId,
		NomeArquivo:   e.NomeArquivo,
		ContentType:   e.ContentType,
		TamanhoBytes:  e.TamanhoBytes,
		Hash:          e.Hash,
		CriadoEm:      e.CriadoEm,
	}
}
