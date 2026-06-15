// Package statement gerencia o carregamento e acesso aos statements SQL
// embarcados para o PostgreSQL via go:embed e sql2go.
package statement

import (
	"embed"
	"strings"

	"github.com/phenpessoa/sql2go"
)

//go:embed *.sql
var source embed.FS

// Statements armazena todos os statements SQL carregados dos arquivos .sql embarcados.
// Cada campo corresponde a um statement nomeado nos arquivos SQL (-- name: NomeQuery).
type Statements struct {
	// usuario
	SalvarUsuario                        string
	BuscarUsuarioPorID                   string
	BuscarUsuarioPorEmail                string
	AtualizarUsuario                     string
	AtualizarUltimoLoginUsuario          string
	SoftDeleteUsuario                    string
	ContarUsuarios                       string
	ContarAdministradoresAtivos          string
	ListarUsuarios                       string
	SalvarUsuarioAuth                    string
	BuscarUsuarioAuthPorProvider         string
	SalvarUsuarioRefreshToken            string
	BuscarUsuarioRefreshTokenPorHash     string
	RevogarUsuarioRefreshToken           string
	RevogarUsuarioRefreshTokensDoUsuario string
	ExpurgarUsuarioRefreshTokens         string
	SalvarUsuarioResetSenhaToken         string
	BuscarUsuarioResetSenhaTokenPorHash  string
	MarcarUsuarioResetSenhaTokenUsado    string

	SalvarUsuarioVerificacaoEmailToken        string
	BuscarUsuarioVerificacaoEmailTokenPorHash string
	MarcarUsuarioVerificacaoEmailTokenUsado   string

	// plano
	BuscarPlanoAtivoPorNegocio string

	// categoria
	SalvarCategoria      string
	BuscarCategoriaPorID string
	AtualizarCategoria   string
	SoftDeleteCategoria  string
	ContarCategorias     string
	ListarCategorias     string

	// negocio
	SalvarNegocio            string
	BuscarNegocioPorID       string
	AtualizarNegocio         string
	SoftDeleteNegocio        string
	ContarNegocios           string
	ListarNegocios           string
	VincularUsuarioNegocio   string
	ListarNegociosPorUsuario string
	VerificarAcessoNegocio   string

	// unidade
	SalvarUnidade            string
	BuscarUnidadePorID       string
	AtualizarUnidade         string
	SoftDeleteUnidade        string
	ContarUnidadesPorNegocio string
	ListarUnidadesPorNegocio string

	// template
	SalvarTemplate              string
	BuscarTemplatePorID         string
	SoftDeleteTemplate          string
	ContarTemplates             string
	ListarTemplates             string
	ContarTemplatesPorCategoria string
	ListarTemplatesPorCategoria string

	// landingpage
	SalvarLandingPage                 string
	BuscarLandingPagePorID            string
	AtualizarLandingPage              string
	SoftDeleteLandingPage             string
	AtualizarAutosaveLandingPage      string
	BloquearLandingPage               string
	DesbloquearLandingPage            string
	VerificarAcessoLandingPage        string
	ContarLandingPagesPorNegocio      string
	AtualizarPublicacaoLandingPage    string
	AtualizarDespublicacaoLandingPage string

	// lead
	SalvarLead               string
	BuscarLeadPorID          string
	BuscarLeadDuplicado      string
	AtualizarStatusLead      string
	AnonimizarLead           string
	AnonimizarLeadsExpirados string

	// lead_aquisicao
	SalvarLeadAquisicao         string
	BuscarLeadAquisicaoPorID    string
	BuscarLeadAquisicaoPorEmail string
	AtualizarFunilLeadAquisicao string

	// exportacao
	SalvarExportacao                          string
	BuscarExportacaoPorID                     string
	ListarExportacoesPorLandingPage           string
	BuscarMaisRecenteExportacaoPorLandingPage string

	// dashboard
	ObterResumoDashboardNegocio            string
	ObterLandingPageMaisUtilizadaDashboard string
	ObterResumoDashboardAdmin              string

	// auditoria
	SalvarAuditoria      string
	BuscarAuditoriaPorID string

	// evento outbox
	SalvarEventoOutbox           string
	MarcarEventoOutboxPublicado  string
	MarcarEventoOutboxFalha      string
	BuscarEventosOutboxPendentes string

	// faturamento / uso / BI
	RegistrarUsoNegocioEvento            string
	ObterFaturamentoMensalPorCompetencia string
	CriarFaturamentoMensal               string
	AdicionarUsoFaturamentoMensal        string
	FecharFaturamentoMensal              string
	IncrementarBiMetricaDiaria           string
}

var stmts Statements

// Load carrega e parseia todos os arquivos .sql embarcados.
// Deve ser chamado uma vez durante a inicialização da conexão.
func Load() error {
	raw, err := loadRaw()
	if err != nil {
		return err
	}
	return sql2go.Parse(&stmts, strings.NewReader(raw))
}

// Get retorna os statements carregados.
func Get() Statements {
	return stmts
}

func loadRaw() (string, error) {
	entries, err := source.ReadDir(".")
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for _, entry := range entries {
		data, err := source.ReadFile(entry.Name())
		if err != nil {
			return "", err
		}
		if len(data) == 0 {
			continue
		}
		b.Write(data)
		b.WriteString("\n\n")
	}
	return b.String(), nil
}
