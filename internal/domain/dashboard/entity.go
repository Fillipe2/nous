package dashboard

import "errors"

var ErrNegocioObrigatorio = errors.New("negocio_id e obrigatorio")

type FiltroPeriodo struct {
	DataInicio    string
	DataFim       string
	NegocioId     string
	UsuarioId     string
	LandingPageId string
	UnidadeId     string
	CategoriaId   string
	TemplateId    string
	Tipo          string
}

type ResumoNegocio struct {
	NegocioId            string
	UsuariosVinculados   int
	LandingPagesCriadas  int
	LandingPagesOnline   int
	ExportacoesGeradas   int
	LandingPageMaisUsada *LandingPageUtilizada
}

type ResumoUsuario struct {
	UsuarioId            string
	NegociosVinculados   int
	LandingPagesCriadas  int
	LandingPagesOnline   int
	ExportacoesGeradas   int
	LandingPageMaisUsada *LandingPageUtilizada
}

type LandingPageUtilizada struct {
	Id               string
	Nome             string
	TotalExportacoes int
}

type ResumoAdmin struct {
	NegociosAtivos         int
	UsuariosAtivos         int
	TemplatesSognix        int
	TemplatesNegocios      int
	LandingPagesCriadas    int
	LandingPagesOnline     int
	LeadsCapturados        int
	ExportacoesGeradas     int
	NegociosComPlanoCustom int
}
