package exportacao

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "nous/internal/domain/exportacao"
	"nous/internal/domain/landingpage"
)

type exportacaoRepositoryFake struct {
	salva      *domain.Exportacao
	exportacao *domain.Exportacao
	filtro     domain.FiltroListagem
}

func (r *exportacaoRepositoryFake) Salvar(_ context.Context, e *domain.Exportacao) error {
	e.Id = "exportacao-1"
	r.salva = e
	return nil
}

func (r *exportacaoRepositoryFake) BuscarPorID(context.Context, string) (*domain.Exportacao, error) {
	if r.exportacao == nil {
		return nil, domain.ErrNaoEncontrada
	}
	return r.exportacao, nil
}

func (r *exportacaoRepositoryFake) Listar(_ context.Context, filtro domain.FiltroListagem) ([]*domain.Exportacao, int, error) {
	r.filtro = filtro
	return []*domain.Exportacao{{Id: "exportacao-1", LandingPageId: filtro.LandingPageId}}, 1, nil
}
func (r *exportacaoRepositoryFake) BuscarMaisRecentePorLandingPage(context.Context, string) (*domain.Exportacao, error) {
	return nil, domain.ErrNaoEncontrada
}

type landingPageRepositoryFake struct{ err error }

func (r *landingPageRepositoryFake) Salvar(context.Context, *landingpage.LandingPage) error {
	return nil
}
func (r *landingPageRepositoryFake) BuscarPorID(context.Context, string) (*landingpage.LandingPage, error) {
	return nil, landingpage.ErrNaoEncontrada
}
func (r *landingPageRepositoryFake) Atualizar(context.Context, *landingpage.LandingPage) error {
	return nil
}
func (r *landingPageRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *landingPageRepositoryFake) Listar(context.Context, landingpage.FiltroListagem) ([]*landingpage.LandingPage, int, error) {
	return nil, 0, nil
}
func (r *landingPageRepositoryFake) ContarPorNegocio(context.Context, string) (int, error) {
	return 0, nil
}
func (r *landingPageRepositoryFake) AtualizarAutosave(context.Context, string, []byte, int) (int, error) {
	return 0, nil
}
func (r *landingPageRepositoryFake) AtualizarLock(context.Context, string, *string) error { return nil }
func (r *landingPageRepositoryFake) VerificarAcessoUsuario(context.Context, string, string) error {
	return r.err
}
func (r *landingPageRepositoryFake) AtualizarPublicacao(context.Context, string, string, string, time.Time) error {
	return nil
}
func (r *landingPageRepositoryFake) AtualizarDespublicacao(context.Context, string) error {
	return nil
}

func TestExportarLandingPageValidaAcessoAntesDeSalvar(t *testing.T) {
	repository := &exportacaoRepositoryFake{}
	useCase := NewUseCase(repository, &landingPageRepositoryFake{err: landingpage.ErrNaoAutorizado})

	_, err := useCase.ExportarLandingPage(context.Background(), "lp-1", "usuario-1", "usuario", "lp.zip", "application/zip", []byte("zip"))

	if !errors.Is(err, domain.ErrNaoAutorizada) {
		t.Fatalf("erro esperado de exportacao nao autorizada, recebido %v", err)
	}
	if repository.salva != nil {
		t.Fatal("exportacao sem acesso nao deveria ser salva")
	}
}

func TestExportarLandingPageDefineContentTypePadraoEHash(t *testing.T) {
	repository := &exportacaoRepositoryFake{}
	useCase := NewUseCase(repository, &landingPageRepositoryFake{})

	output, err := useCase.ExportarLandingPage(context.Background(), "lp-1", "usuario-1", "usuario", "lp.zip", "", []byte("zip"))

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.ContentType != "application/zip" || output.Hash == "" || output.TamanhoBytes != 3 {
		t.Fatalf("saida inesperada: %+v", output)
	}
	if repository.salva == nil || repository.salva.Hash != output.Hash {
		t.Fatalf("exportacao salva inesperada: %+v", repository.salva)
	}
}

func TestListarExportacoesExigeLandingPageParaUsuario(t *testing.T) {
	useCase := NewUseCase(&exportacaoRepositoryFake{}, &landingPageRepositoryFake{})

	_, err := useCase.ListarExportacoes(context.Background(), domain.FiltroListagem{}, "usuario-1", "usuario")

	if err == nil {
		t.Fatal("erro esperado quando usuario lista sem landingpage_id")
	}
}

func TestBaixarExportacaoValidaAcessoDaLandingPage(t *testing.T) {
	useCase := NewUseCase(&exportacaoRepositoryFake{exportacao: &domain.Exportacao{Id: "exportacao-1", LandingPageId: "lp-1", Arquivo: []byte("zip")}}, &landingPageRepositoryFake{err: landingpage.ErrNaoAutorizado})

	_, _, err := useCase.BaixarExportacao(context.Background(), "exportacao-1", "usuario-1", "usuario")

	if !errors.Is(err, domain.ErrNaoAutorizada) {
		t.Fatalf("erro esperado de acesso, recebido %v", err)
	}
}
