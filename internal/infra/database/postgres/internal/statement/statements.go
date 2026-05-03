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
	SoftDeleteUsuario                    string
	ContarUsuarios                       string
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

	// plano
	BuscarPlanoAtivoPorCliente string

	// categoria
	SalvarCategoria      string
	BuscarCategoriaPorID string
	AtualizarCategoria   string
	SoftDeleteCategoria  string
	ContarCategorias     string
	ListarCategorias     string

	// cliente
	SalvarCliente            string
	BuscarClientePorID       string
	AtualizarCliente         string
	SoftDeleteCliente        string
	ContarClientes           string
	ListarClientes           string
	VincularUsuarioCliente   string
	ListarClientesPorUsuario string
	VerificarAcessoCliente   string

	// negocio
	SalvarNegocio            string
	BuscarNegocioPorID       string
	AtualizarNegocio         string
	SoftDeleteNegocio        string
	ContarNegociosPorCliente string
	ListarNegociosPorCliente string

	// template
	SalvarTemplate              string
	BuscarTemplatePorID         string
	SoftDeleteTemplate          string
	ContarTemplates             string
	ListarTemplates             string
	ContarTemplatesPorCategoria string
	ListarTemplatesPorCategoria string

	// landingpage
	SalvarLandingPage            string
	BuscarLandingPagePorID       string
	AtualizarLandingPage         string
	SoftDeleteLandingPage        string
	AtualizarAutosaveLandingPage string
	BloquearLandingPage          string
	DesbloquearLandingPage       string
	VerificarAcessoLandingPage   string
	ContarLandingPagesPorCliente string

	// lead
	SalvarLead               string
	BuscarLeadPorID          string
	BuscarLeadDuplicado      string
	AtualizarStatusLead      string
	AnonimizarLead           string
	AnonimizarLeadsExpirados string

	// exportacao
	SalvarExportacao                string
	BuscarExportacaoPorID           string
	ListarExportacoesPorLandingPage string

	// dashboard
	ObterResumoDashboardCliente            string
	ObterLandingPageMaisUtilizadaDashboard string
	ObterResumoDashboardAdmin              string

	// auditoria
	SalvarAuditoria      string
	BuscarAuditoriaPorID string

	// evento outbox
	SalvarEventoOutbox          string
	MarcarEventoOutboxPublicado string
	MarcarEventoOutboxFalha     string
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
