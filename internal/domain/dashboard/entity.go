package dashboard

import "errors"

var ErrClienteObrigatorio = errors.New("cliente_id e obrigatorio")

type FiltroPeriodo struct {
	DataInicio string
	DataFim    string
}

type ResumoCliente struct {
	ClienteId            string
	UsuariosVinculados   int
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
	ClientesAtivos         int
	UsuariosAtivos         int
	TemplatesSognix        int
	TemplatesClientes      int
	LandingPagesCriadas    int
	LandingPagesOnline     int
	LeadsCapturados        int
	ExportacoesGeradas     int
	ClientesComPlanoCustom int
}
