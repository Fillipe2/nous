package dto

type DashboardClienteOutput struct {
	ClienteId            string                      `json:"cliente_id"`
	UsuariosVinculados   int                         `json:"usuarios_vinculados"`
	LandingPagesCriadas  int                         `json:"landingpages_criadas"`
	LandingPagesOnline   int                         `json:"landingpages_online"`
	ExportacoesGeradas   int                         `json:"exportacoes_geradas"`
	LandingPageMaisUsada *LandingPageUtilizadaOutput `json:"landingpage_mais_usada,omitempty"`
}

type DashboardUsuarioOutput struct {
	UsuarioId            string                      `json:"usuario_id"`
	ClientesVinculados   int                         `json:"clientes_vinculados"`
	LandingPagesCriadas  int                         `json:"landingpages_criadas"`
	LandingPagesOnline   int                         `json:"landingpages_online"`
	ExportacoesGeradas   int                         `json:"exportacoes_geradas"`
	LandingPageMaisUsada *LandingPageUtilizadaOutput `json:"landingpage_mais_usada,omitempty"`
}

type LandingPageUtilizadaOutput struct {
	Id               string `json:"id"`
	Nome             string `json:"nome"`
	TotalExportacoes int    `json:"total_exportacoes"`
}

type DashboardAdminOutput struct {
	ClientesAtivos         int `json:"clientes_ativos"`
	UsuariosAtivos         int `json:"usuarios_ativos"`
	TemplatesSognix        int `json:"templates_sognix"`
	TemplatesClientes      int `json:"templates_clientes"`
	LandingPagesCriadas    int `json:"landingpages_criadas"`
	LandingPagesOnline     int `json:"landingpages_online"`
	LeadsCapturados        int `json:"leads_capturados"`
	ExportacoesGeradas     int `json:"exportacoes_geradas"`
	ClientesComPlanoCustom int `json:"clientes_com_plano_customizavel"`
}
