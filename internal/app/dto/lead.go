package dto

import (
	"encoding/json"
	"time"
)

type CapturarLeadInput struct {
	LandingPageId                   string          `json:"landingpage_id"`
	Nome                            string          `json:"nome"`
	Email                           string          `json:"email"`
	Telefone                        string          `json:"telefone,omitempty"`
	Whatsapp                        string          `json:"whatsapp,omitempty"`
	Mensagem                        string          `json:"mensagem,omitempty"`
	CamposPersonalizados            json.RawMessage `json:"campos_personalizados,omitempty"`
	Consentimento                   bool            `json:"consentimento"`
	PoliticaVersao                  string          `json:"politica_versao,omitempty"`
	ConsentimentoMarketingParceiros bool            `json:"consentimento_marketing_parceiros,omitempty"`
	ConsentimentoVersaoMarketing    string          `json:"consentimento_versao_marketing,omitempty"`
	UTM                             *UTMInput       `json:"utm,omitempty"`
	Referrer                        string          `json:"referrer,omitempty"`
	// Honeypot: campo escondido no formulario; bots tendem a preencher.
	// Submissoes com este campo nao vazio sao silenciosamente descartadas.
	Honeypot string `json:"website,omitempty"`
}

type UTMInput struct {
	Source   string `json:"source,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Campaign string `json:"campaign,omitempty"`
	Content  string `json:"content,omitempty"`
	Term     string `json:"term,omitempty"`
}

type AtualizarStatusLeadInput struct {
	Status string `json:"status"`
	Motivo string `json:"motivo,omitempty"`
}

type LeadOutput struct {
	Id                   string          `json:"id"`
	LandingPageId        string          `json:"landingpage_id"`
	LandingPageNome      string          `json:"landingpage_nome,omitempty"`
	ClienteId            string          `json:"cliente_id,omitempty"`
	Nome                 string          `json:"nome"`
	Email                string          `json:"email"`
	Telefone             string          `json:"telefone"`
	Whatsapp             string          `json:"whatsapp"`
	Mensagem             string          `json:"mensagem"`
	CamposPersonalizados json.RawMessage `json:"campos_personalizados,omitempty"`
	Origem               string          `json:"origem"`
	Status               string          `json:"status"`
	PerdidoMotivo        string          `json:"perdido_motivo,omitempty"`
	UTM                  *UTMInput       `json:"utm,omitempty"`
	Referrer             string          `json:"referrer,omitempty"`
	Consentimento                   bool            `json:"consentimento"`
	ConsentimentoEm                 *time.Time      `json:"consentimento_em,omitempty"`
	PoliticaVersao                  string          `json:"politica_versao,omitempty"`
	ConsentimentoMarketingParceiros bool            `json:"consentimento_marketing_parceiros,omitempty"`
	ConsentimentoVersaoMarketing    string          `json:"consentimento_versao_marketing,omitempty"`
	AnonimizadoEm                   *time.Time      `json:"anonimizado_em,omitempty"`
	CriadoEm             time.Time       `json:"criado_em"`
}

type ListaLeadsOutput struct {
	Dados     []*LeadOutput `json:"dados"`
	Total     int           `json:"total"`
	Pagina    int           `json:"pagina"`
	PorPagina int           `json:"por_pagina"`
}
