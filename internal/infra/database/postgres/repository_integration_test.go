package postgres

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"buildnix-backend/internal/domain/categoria"
	"buildnix-backend/internal/domain/cliente"
	"buildnix-backend/internal/domain/dashboard"
	"buildnix-backend/internal/domain/exportacao"
	"buildnix-backend/internal/domain/landingpage"
	"buildnix-backend/internal/domain/lead"
	"buildnix-backend/internal/domain/negocio"
	"buildnix-backend/internal/domain/template"
	"buildnix-backend/internal/domain/usuario"
)

func TestPostgresRepositoriesIntegrationBasico(t *testing.T) {
	ctx := context.Background()
	db := conectarPostgresIntegracao(t, ctx)

	usuarioRepo, err := NewUsuarioRepository(db)
	if err != nil {
		t.Fatalf("criar usuario repo: %v", err)
	}
	clienteRepo, err := NewClienteRepository(db)
	if err != nil {
		t.Fatalf("criar cliente repo: %v", err)
	}
	categoriaRepo, err := NewCategoriaRepository(db)
	if err != nil {
		t.Fatalf("criar categoria repo: %v", err)
	}
	negocioRepo, err := NewNegocioRepository(db)
	if err != nil {
		t.Fatalf("criar negocio repo: %v", err)
	}
	templateRepo, err := NewTemplateRepository(db)
	if err != nil {
		t.Fatalf("criar template repo: %v", err)
	}
	landingPageRepo, err := NewLandingPageRepository(db)
	if err != nil {
		t.Fatalf("criar landingpage repo: %v", err)
	}
	leadRepo, err := NewLeadRepository(db)
	if err != nil {
		t.Fatalf("criar lead repo: %v", err)
	}
	exportacaoRepo, err := NewExportacaoRepository(db)
	if err != nil {
		t.Fatalf("criar exportacao repo: %v", err)
	}
	dashboardRepo, err := NewDashboardRepository(db)
	if err != nil {
		t.Fatalf("criar dashboard repo: %v", err)
	}

	agora := time.Now().UTC()
	us := &usuario.Usuario{Nome: "Usuario Integracao", Email: "usuario.integracao@sognix.test", SenhaHash: "hash", Role: usuario.RoleUsuario, Status: usuario.StatusAtivo, CriadoEm: agora, AtualizadoEm: agora, Version: 1}
	if err := usuarioRepo.Salvar(ctx, us); err != nil {
		t.Fatalf("salvar usuario: %v", err)
	}
	usuarioPersistido, err := usuarioRepo.BuscarPorEmail(ctx, us.Email)
	if err != nil {
		t.Fatalf("buscar usuario por email: %v", err)
	}
	if usuarioPersistido.Id != us.Id || usuarioPersistido.SenhaHash != "hash" {
		t.Fatalf("usuario persistido inesperado: %+v", usuarioPersistido)
	}

	cl := &cliente.Cliente{Nome: "Cliente Integracao", Email: "cliente@sognix.test", Status: cliente.StatusAtivo, CriadoEm: agora, AtualizadoEm: agora, Version: 1}
	if err := clienteRepo.Salvar(ctx, cl); err != nil {
		t.Fatalf("salvar cliente: %v", err)
	}
	if err := clienteRepo.VincularUsuario(ctx, &cliente.UsuarioCliente{UsuarioId: us.Id, ClienteId: cl.Id, CriadoEm: agora}); err != nil {
		t.Fatalf("vincular usuario cliente: %v", err)
	}
	if err := clienteRepo.VerificarAcessoUsuario(ctx, cl.Id, us.Id); err != nil {
		t.Fatalf("verificar acesso cliente: %v", err)
	}

	cat := &categoria.Categoria{Nome: "Categoria Integracao", Status: categoria.StatusAtivo, CriadoEm: agora, AtualizadoEm: agora, Version: 1}
	if err := categoriaRepo.Salvar(ctx, cat); err != nil {
		t.Fatalf("salvar categoria: %v", err)
	}

	neg := &negocio.Negocio{ClienteId: cl.Id, Nome: "Negocio Integracao", Status: negocio.StatusAtivo, CriadoEm: agora, AtualizadoEm: agora, Version: 1}
	if err := negocioRepo.Salvar(ctx, neg); err != nil {
		t.Fatalf("salvar negocio: %v", err)
	}

	tpl := &template.Template{ClienteId: cl.Id, CategoriaId: cat.Id, Tipo: template.TipoWhiteLabel, Origem: template.OrigemCliente, Nome: "Template Integracao", Conteudo: "<html></html>", Status: template.StatusAtivo, CriadoEm: agora, AtualizadoEm: agora, Version: 1}
	if err := templateRepo.Salvar(ctx, tpl); err != nil {
		t.Fatalf("salvar template: %v", err)
	}
	templates, totalTemplates, err := templateRepo.Listar(ctx, template.FiltroListagem{ClienteId: cl.Id, Pagina: 1, PorPagina: 10})
	if err != nil {
		t.Fatalf("listar templates: %v", err)
	}
	if totalTemplates != 1 || len(templates) != 1 {
		t.Fatalf("templates esperados=1, total=%d len=%d", totalTemplates, len(templates))
	}

	lp := &landingpage.LandingPage{ClienteId: cl.Id, NegocioId: neg.Id, CategoriaId: cat.Id, TemplateId: tpl.Id, Tipo: landingpage.TipoWhiteLabel, Nome: "Landing Integracao", Conteudo: "html", Status: landingpage.StatusRascunho, IsDraft: true, AutosaveData: json.RawMessage(`{"draft":true}`), CriadoEm: agora, AtualizadoEm: agora, Version: 1}
	if err := landingPageRepo.Salvar(ctx, lp); err != nil {
		t.Fatalf("salvar landingpage: %v", err)
	}
	if total, err := landingPageRepo.ContarPorCliente(ctx, cl.Id); err != nil || total != 1 {
		t.Fatalf("contar landingpages: total=%d err=%v", total, err)
	}
	if err := landingPageRepo.VerificarAcessoUsuario(ctx, lp.Id, us.Id); err != nil {
		t.Fatalf("verificar acesso landingpage: %v", err)
	}
	novaVersao, err := landingPageRepo.AtualizarAutosave(ctx, lp.Id, []byte(`{"draft":false}`), 1)
	if err != nil || novaVersao != 2 {
		t.Fatalf("autosave: versao=%d err=%v", novaVersao, err)
	}

	le := &lead.Lead{LandingPageId: lp.Id, Nome: "Lead Integracao", Email: "lead@sognix.test", Origem: "teste", Status: lead.StatusNovo, CriadoEm: agora}
	if err := leadRepo.Salvar(ctx, le); err != nil {
		t.Fatalf("salvar lead: %v", err)
	}
	if err := leadRepo.AtualizarStatus(ctx, le.Id, lead.AtualizacaoStatus{Status: lead.StatusConvertido}, agora); err != nil {
		t.Fatalf("atualizar status lead: %v", err)
	}

	exp := &exportacao.Exportacao{LandingPageId: lp.Id, UsuarioId: us.Id, NomeArquivo: "landing.zip", ContentType: "application/zip", TamanhoBytes: 3, Arquivo: []byte("zip"), Hash: "hash-exportacao", CriadoEm: agora}
	if err := exportacaoRepo.Salvar(ctx, exp); err != nil {
		t.Fatalf("salvar exportacao: %v", err)
	}
	exportacaoPersistida, err := exportacaoRepo.BuscarPorID(ctx, exp.Id)
	if err != nil {
		t.Fatalf("buscar exportacao: %v", err)
	}
	if string(exportacaoPersistida.Arquivo) != "zip" || exportacaoPersistida.Hash != "hash-exportacao" {
		t.Fatalf("exportacao persistida inesperada: %+v", exportacaoPersistida)
	}

	resumoAdmin, err := dashboardRepo.ObterResumoAdmin(ctx, dashboard.FiltroPeriodo{ClienteId: cl.Id, UsuarioId: us.Id, LandingPageId: lp.Id, DataInicio: agora.Format("2006-01-02"), DataFim: agora.Format("2006-01-02")})
	if err != nil {
		t.Fatalf("obter dashboard admin filtrado: %v", err)
	}
	if resumoAdmin.ClientesAtivos != 1 || resumoAdmin.UsuariosAtivos != 1 || resumoAdmin.LandingPagesCriadas != 1 || resumoAdmin.LeadsCapturados != 1 || resumoAdmin.ExportacoesGeradas != 1 {
		t.Fatalf("dashboard admin filtrado inesperado: %+v", resumoAdmin)
	}

	resumoUsuario, err := dashboardRepo.ObterResumoUsuario(ctx, us.Id, dashboard.FiltroPeriodo{ClienteId: cl.Id, LandingPageId: lp.Id})
	if err != nil {
		t.Fatalf("obter dashboard usuario filtrado: %v", err)
	}
	if resumoUsuario.LandingPagesCriadas != 1 || resumoUsuario.ExportacoesGeradas != 1 {
		t.Fatalf("dashboard usuario filtrado inesperado: %+v", resumoUsuario)
	}
}

func TestUsuarioRefreshTokenRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	db := conectarPostgresIntegracao(t, ctx)
	repo, err := NewUsuarioRepository(db)
	if err != nil {
		t.Fatalf("criar usuario repo: %v", err)
	}

	agora := time.Now().UTC()
	us := &usuario.Usuario{Nome: "Usuario Refresh", Email: "refresh@sognix.test", SenhaHash: "hash", Role: usuario.RoleUsuario, Status: usuario.StatusAtivo, CriadoEm: agora, AtualizadoEm: agora, Version: 1}
	if err := repo.Salvar(ctx, us); err != nil {
		t.Fatalf("salvar usuario: %v", err)
	}
	token := &usuario.RefreshToken{UsuarioId: us.Id, TokenHash: strings.Repeat("a", 64), ExpiresAt: agora.Add(time.Hour), CriadoEm: agora}
	if err := repo.SalvarRefreshToken(ctx, token); err != nil {
		t.Fatalf("salvar refresh token: %v", err)
	}
	persistido, err := repo.BuscarRefreshTokenPorHash(ctx, token.TokenHash)
	if err != nil {
		t.Fatalf("buscar refresh token: %v", err)
	}
	if persistido.UsuarioId != us.Id || persistido.TokenHash != token.TokenHash {
		t.Fatalf("refresh persistido inesperado: %+v", persistido)
	}
	if err := repo.RevogarRefreshToken(ctx, token.TokenHash, agora); err != nil {
		t.Fatalf("revogar refresh token: %v", err)
	}
	revogado, err := repo.BuscarRefreshTokenPorHash(ctx, token.TokenHash)
	if err != nil {
		t.Fatalf("buscar refresh revogado: %v", err)
	}
	if revogado.RevogadoEm == nil {
		t.Fatal("refresh token deveria estar revogado")
	}

	tokenNovo := &usuario.RefreshToken{UsuarioId: us.Id, TokenHash: strings.Repeat("b", 64), ExpiresAt: agora.Add(time.Hour), CriadoEm: agora}
	if err := repo.SalvarRefreshToken(ctx, tokenNovo); err != nil {
		t.Fatalf("salvar refresh token novo: %v", err)
	}
	if err := repo.RevogarRefreshTokensDoUsuario(ctx, us.Id, agora); err != nil {
		t.Fatalf("revogar refresh tokens do usuario: %v", err)
	}
	revogadoPorUsuario, err := repo.BuscarRefreshTokenPorHash(ctx, tokenNovo.TokenHash)
	if err != nil {
		t.Fatalf("buscar refresh revogado por usuario: %v", err)
	}
	if revogadoPorUsuario.RevogadoEm == nil {
		t.Fatal("refresh token deveria ser revogado pelo usuario")
	}
}

func conectarPostgresIntegracao(t *testing.T, ctx context.Context) *Postgres {
	t.Helper()
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("defina TEST_DATABASE_URL para rodar testes de integracao Postgres")
	}

	schema := "buildnix_test_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	basePool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("conectar postgres base: %v", err)
	}
	if _, err := basePool.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		basePool.Close()
		t.Fatalf("criar schema temporario: %v", err)
	}
	t.Cleanup(func() {
		_, _ = basePool.Exec(context.Background(), `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
		basePool.Close()
	})

	db, err := Connect(ctx, databaseURLComSearchPath(t, databaseURL, schema))
	if err != nil {
		t.Fatalf("connect postgres schema temporario: %v", err)
	}
	t.Cleanup(db.Close)
	aplicarMigracoesIntegracao(t, ctx, db)
	return db
}

func aplicarMigracoesIntegracao(t *testing.T, ctx context.Context, db *Postgres) {
	t.Helper()
	_, arquivo, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("nao foi possivel localizar arquivo de teste")
	}
	raiz := filepath.Clean(filepath.Join(filepath.Dir(arquivo), "../../../.."))
	migracoes, err := filepath.Glob(filepath.Join(raiz, "migrations/postgres/*.up.sql"))
	if err != nil {
		t.Fatalf("listar migracoes: %v", err)
	}
	if len(migracoes) == 0 {
		t.Fatal("nenhuma migracao encontrada")
	}
	for _, migracao := range migracoes {
		conteudo, err := os.ReadFile(migracao)
		if err != nil {
			t.Fatalf("ler migracao %s: %v", filepath.Base(migracao), err)
		}
		if err := db.Exec(ctx, string(conteudo)); err != nil {
			t.Fatalf("aplicar migracao %s: %v", filepath.Base(migracao), err)
		}
	}
}

func databaseURLComSearchPath(t *testing.T, databaseURL, schema string) string {
	t.Helper()
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
