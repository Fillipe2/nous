package lead_aquisicao

import (
	"context"
	"errors"
	"testing"
	"time"

	"buildnix-backend/internal/app/dto"
	"buildnix-backend/internal/domain/evento"
	domain "buildnix-backend/internal/domain/lead_aquisicao"
)

// --- fakes ---

type repositoryFake struct {
	salvo      *domain.LeadAquisicao
	porEmail   *domain.LeadAquisicao
	porID      *domain.LeadAquisicao
	funil      domain.AtualizacaoFunil
	funilID    string
	listar     []*domain.LeadAquisicao
	erroSalvar error
}

func (r *repositoryFake) Salvar(_ context.Context, la *domain.LeadAquisicao) error {
	if r.erroSalvar != nil {
		return r.erroSalvar
	}
	la.Id = "la-test"
	r.salvo = la
	return nil
}

func (r *repositoryFake) BuscarPorID(_ context.Context, id string) (*domain.LeadAquisicao, error) {
	if r.porID == nil {
		return nil, domain.ErrNaoEncontrado
	}
	return r.porID, nil
}

func (r *repositoryFake) BuscarPorEmail(_ context.Context, _ string) (*domain.LeadAquisicao, error) {
	if r.porEmail == nil {
		return nil, domain.ErrNaoEncontrado
	}
	return r.porEmail, nil
}

func (r *repositoryFake) AtualizarFunil(_ context.Context, id string, at domain.AtualizacaoFunil, _ time.Time) error {
	r.funilID = id
	r.funil = at
	return nil
}

func (r *repositoryFake) Listar(_ context.Context, _ domain.FiltroListagem) ([]*domain.LeadAquisicao, int, error) {
	return r.listar, len(r.listar), nil
}

type publisherFake struct {
	publicado bool
	ultimo    evento.Evento
	err       error
}

func (p *publisherFake) Publicar(_ context.Context, e evento.Evento) error {
	p.publicado = true
	p.ultimo = e
	return p.err
}

// --- helpers ---

func inputValido() dto.RegistrarLeadAquisicaoInput {
	return dto.RegistrarLeadAquisicaoInput{
		Nome:  "Dr. João Silva",
		Email: "joao@clinica.com.br",
		Porte: "pequeno",
	}
}

func novoUseCase(repo *repositoryFake, pub *publisherFake) *UseCase {
	return NewUseCase(repo, pub)
}

// --- testes de RegistrarLeadAquisicao ---

func TestRegistrarLeadAquisicaoSalvaComSucesso(t *testing.T) {
	repo := &repositoryFake{}
	pub := &publisherFake{}
	uc := novoUseCase(repo, pub)

	out, err := uc.RegistrarLeadAquisicao(context.Background(), inputValido(), "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Id != "la-test" {
		t.Fatalf("id esperado la-test, recebido %q", out.Id)
	}
	if repo.salvo == nil {
		t.Fatal("lead_aquisicao deveria ter sido salvo")
	}
	if repo.salvo.StatusFunil != domain.StatusFunilNovo {
		t.Fatalf("status funil deveria ser novo, recebido %q", repo.salvo.StatusFunil)
	}
}

func TestRegistrarLeadAquisicaoNormalizaEmailEEstado(t *testing.T) {
	repo := &repositoryFake{}
	uc := novoUseCase(repo, &publisherFake{})

	in := inputValido()
	in.Email = "  JOAO@CLINICA.COM.BR  "
	in.Estado = "sp"

	_, err := uc.RegistrarLeadAquisicao(context.Background(), in, "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.salvo.Email != "joao@clinica.com.br" {
		t.Fatalf("email nao normalizado: %q", repo.salvo.Email)
	}
	if repo.salvo.Estado != "SP" {
		t.Fatalf("estado nao normalizado para maiusculo: %q", repo.salvo.Estado)
	}
}

func TestRegistrarLeadAquisicaoRejeitaEmailInvalido(t *testing.T) {
	uc := novoUseCase(&repositoryFake{}, &publisherFake{})

	in := inputValido()
	in.Email = "nao-e-email"
	_, err := uc.RegistrarLeadAquisicao(context.Background(), in, "127.0.0.1", "test")
	if err == nil {
		t.Fatal("esperado erro para email invalido")
	}
}

func TestRegistrarLeadAquisicaoRejeitaNomeVazio(t *testing.T) {
	uc := novoUseCase(&repositoryFake{}, &publisherFake{})

	in := inputValido()
	in.Nome = "   "
	_, err := uc.RegistrarLeadAquisicao(context.Background(), in, "127.0.0.1", "test")
	if err == nil {
		t.Fatal("esperado erro para nome vazio")
	}
}

func TestRegistrarLeadAquisicaoRejeitaPorteInvalido(t *testing.T) {
	uc := novoUseCase(&repositoryFake{}, &publisherFake{})

	in := inputValido()
	in.Porte = "gigante"
	_, err := uc.RegistrarLeadAquisicao(context.Background(), in, "127.0.0.1", "test")
	if err == nil {
		t.Fatal("esperado erro para porte invalido")
	}
}

func TestRegistrarLeadAquisicaoRejeitaMensagemLonga(t *testing.T) {
	uc := novoUseCase(&repositoryFake{}, &publisherFake{})

	in := inputValido()
	for i := 0; i <= limiteMensagem; i++ {
		in.Mensagem += "x"
	}
	_, err := uc.RegistrarLeadAquisicao(context.Background(), in, "127.0.0.1", "test")
	if err == nil {
		t.Fatal("esperado erro para mensagem muito longa")
	}
}

func TestRegistrarLeadAquisicaoRejeitaHoneypotPreenchido(t *testing.T) {
	repo := &repositoryFake{}
	uc := novoUseCase(repo, &publisherFake{})

	in := inputValido()
	in.Honeypot = "http://spam.example"
	_, err := uc.RegistrarLeadAquisicao(context.Background(), in, "127.0.0.1", "bot")
	if !errors.Is(err, domain.ErrSpamDetectado) {
		t.Fatalf("esperado ErrSpamDetectado, recebido %v", err)
	}
	if repo.salvo != nil {
		t.Fatal("submissao de spam nao deveria ser salva")
	}
}

func TestRegistrarLeadAquisicaoRejeitaEmailDuplicado(t *testing.T) {
	existente := &domain.LeadAquisicao{Id: "outro", Email: "joao@clinica.com.br"}
	repo := &repositoryFake{porEmail: existente}
	uc := novoUseCase(repo, &publisherFake{})

	_, err := uc.RegistrarLeadAquisicao(context.Background(), inputValido(), "127.0.0.1", "test")
	if !errors.Is(err, domain.ErrEmailDuplicado) {
		t.Fatalf("esperado ErrEmailDuplicado, recebido %v", err)
	}
}

func TestRegistrarLeadAquisicaoPublicaEvento(t *testing.T) {
	pub := &publisherFake{}
	uc := novoUseCase(&repositoryFake{}, pub)

	_, err := uc.RegistrarLeadAquisicao(context.Background(), inputValido(), "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !pub.publicado {
		t.Fatal("evento lead_aquisicao.registrado deveria ter sido publicado")
	}
	if pub.ultimo.EventType != "lead_aquisicao.registrado" {
		t.Fatalf("event_type incorreto: %q", pub.ultimo.EventType)
	}
	if pub.ultimo.Payload["lead_aquisicao_id"] != "la-test" {
		t.Fatalf("payload sem lead_aquisicao_id correto: %+v", pub.ultimo.Payload)
	}
}

func TestRegistrarLeadAquisicaoIgnoraErroDePublicacao(t *testing.T) {
	pub := &publisherFake{err: errors.New("observer indisponivel")}
	repo := &repositoryFake{}
	uc := novoUseCase(repo, pub)

	out, err := uc.RegistrarLeadAquisicao(context.Background(), inputValido(), "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("erro do publisher nao deveria interromper o registro: %v", err)
	}
	if out == nil || repo.salvo == nil {
		t.Fatal("lead_aquisicao deveria ter sido salvo mesmo com erro de publicacao")
	}
}

func TestRegistrarLeadAquisicaoPreencheUTM(t *testing.T) {
	repo := &repositoryFake{}
	uc := novoUseCase(repo, &publisherFake{})

	in := inputValido()
	in.UTM = &dto.UTMInput{Source: "google", Medium: "cpc", Campaign: "lancamento_q2"}

	_, err := uc.RegistrarLeadAquisicao(context.Background(), in, "10.0.0.1", "test")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.salvo.UTMSource != "google" || repo.salvo.UTMCampaign != "lancamento_q2" {
		t.Fatalf("UTM nao persistido corretamente: source=%q campaign=%q", repo.salvo.UTMSource, repo.salvo.UTMCampaign)
	}
	if repo.salvo.IP != "10.0.0.1" {
		t.Fatalf("IP nao persistido: %q", repo.salvo.IP)
	}
}

// --- testes de AtualizarFunilLeadAquisicao ---

func TestAtualizarFunilValidaStatus(t *testing.T) {
	uc := novoUseCase(&repositoryFake{}, &publisherFake{})

	err := uc.AtualizarFunilLeadAquisicao(context.Background(), "id-1", dto.AtualizarFunilLeadAquisicaoInput{StatusFunil: "invalido"})
	if !errors.Is(err, domain.ErrStatusInvalido) {
		t.Fatalf("esperado ErrStatusInvalido, recebido %v", err)
	}
}

func TestAtualizarFunilPersiste(t *testing.T) {
	repo := &repositoryFake{}
	uc := novoUseCase(repo, &publisherFake{})

	err := uc.AtualizarFunilLeadAquisicao(context.Background(), "id-1", dto.AtualizarFunilLeadAquisicaoInput{
		StatusFunil:           string(domain.StatusFunilConvertido),
		ConvertidoEmUsuarioId: "usuario-abc",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.funilID != "id-1" {
		t.Fatalf("id nao propagado: %q", repo.funilID)
	}
	if repo.funil.StatusFunil != domain.StatusFunilConvertido {
		t.Fatalf("status nao propagado: %q", repo.funil.StatusFunil)
	}
	if repo.funil.ConvertidoEmUsuarioId != "usuario-abc" {
		t.Fatalf("convertido_em nao propagado: %q", repo.funil.ConvertidoEmUsuarioId)
	}
}

// --- testes de ObterLeadAquisicao ---

func TestObterLeadAquisicaoNaoEncontrado(t *testing.T) {
	uc := novoUseCase(&repositoryFake{}, &publisherFake{})

	_, err := uc.ObterLeadAquisicao(context.Background(), "inexistente")
	if err == nil {
		t.Fatal("esperado erro para lead nao encontrado")
	}
}

func TestObterLeadAquisicaoRetornaOutput(t *testing.T) {
	la := &domain.LeadAquisicao{Id: "la-1", Nome: "Teste", Email: "t@t.com", StatusFunil: domain.StatusFunilNovo, CriadoEm: time.Now(), AtualizadoEm: time.Now()}
	repo := &repositoryFake{porID: la}
	uc := novoUseCase(repo, &publisherFake{})

	out, err := uc.ObterLeadAquisicao(context.Background(), "la-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Id != "la-1" || out.Nome != "Teste" {
		t.Fatalf("output incorreto: %+v", out)
	}
}

// --- testes de ListarLeadsAquisicao ---

func TestListarLeadsAquisicaoRetornaVazioSemItens(t *testing.T) {
	uc := novoUseCase(&repositoryFake{}, &publisherFake{})

	out, err := uc.ListarLeadsAquisicao(context.Background(), domain.FiltroListagem{Pagina: 1, PorPagina: 20})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(out.Dados) != 0 || out.Total != 0 {
		t.Fatalf("esperado lista vazia, recebido %+v", out)
	}
}

func TestListarLeadsAquisicaoRetornaItens(t *testing.T) {
	itens := []*domain.LeadAquisicao{
		{Id: "a", Nome: "Ana", Email: "ana@test.com", StatusFunil: domain.StatusFunilNovo, CriadoEm: time.Now(), AtualizadoEm: time.Now()},
		{Id: "b", Nome: "Bruno", Email: "bruno@test.com", StatusFunil: domain.StatusFunilContatado, CriadoEm: time.Now(), AtualizadoEm: time.Now()},
	}
	repo := &repositoryFake{listar: itens}
	uc := novoUseCase(repo, &publisherFake{})

	out, err := uc.ListarLeadsAquisicao(context.Background(), domain.FiltroListagem{Pagina: 1, PorPagina: 20})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Total != 2 || len(out.Dados) != 2 {
		t.Fatalf("esperado 2 itens, recebido total=%d dados=%d", out.Total, len(out.Dados))
	}
}
