package dto

type DashboardNegocioOutput struct {
	NegocioId            string                      `json:"negocio_id"`
	UsuariosVinculados   int                         `json:"usuarios_vinculados"`
	LandingPagesCriadas  int                         `json:"landingpages_criadas"`
	LandingPagesOnline   int                         `json:"landingpages_online"`
	ExportacoesGeradas   int                         `json:"exportacoes_geradas"`
	LandingPageMaisUsada *LandingPageUtilizadaOutput `json:"landingpage_mais_usada,omitempty"`
}

type DashboardUsuarioOutput struct {
	UsuarioId            string                      `json:"usuario_id"`
	NegociosVinculados   int                         `json:"negocios_vinculados"`
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
	NegociosAtivos         int `json:"negocios_ativos"`
	UsuariosAtivos         int `json:"usuarios_ativos"`
	TemplatesSognix        int `json:"templates_sognix"`
	TemplatesNegocios      int `json:"templates_negocios"`
	LandingPagesCriadas    int `json:"landingpages_criadas"`
	LandingPagesOnline     int `json:"landingpages_online"`
	LeadsCapturados        int `json:"leads_capturados"`
	ExportacoesGeradas     int `json:"exportacoes_geradas"`
	NegociosComPlanoCustom int `json:"negocios_com_plano_customizavel"`
}
