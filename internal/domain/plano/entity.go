package plano

import "errors"

var ErrAssinaturaAtivaNaoEncontrada = errors.New("assinatura ativa nao encontrada")

type Plano struct {
	Id                  string
	Nome                string
	Descricao           string
	ValorMensalCentavos int
	Limites             Limites
	Status              string
}

type Limites struct {
	LandingPagesWhiteLabel    *bool `json:"landingpages_white_label,omitempty"`
	LandingPagesCustomizaveis bool  `json:"landingpages_customizaveis,omitempty"`
	MaxLandingPages           int   `json:"max_landingpages,omitempty"`
	MaxExportacoesMes         int   `json:"max_exportacoes_mes,omitempty"`
}

func (l Limites) PermiteWhiteLabel() bool {
	return l.LandingPagesWhiteLabel == nil || *l.LandingPagesWhiteLabel
}

func (l Limites) PermiteCustomizavel() bool {
	return l.LandingPagesCustomizaveis
}
