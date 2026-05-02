package dto

import (
	"encoding/json"
	"time"
)

type CapturarLeadInput struct {
	LandingPageId        string          `json:"landingpage_id"`
	Nome                 string          `json:"nome"`
	Email                string          `json:"email"`
	Telefone             string          `json:"telefone,omitempty"`
	Whatsapp             string          `json:"whatsapp,omitempty"`
	Mensagem             string          `json:"mensagem,omitempty"`
	CamposPersonalizados json.RawMessage `json:"campos_personalizados,omitempty"`
}

type AtualizarStatusLeadInput struct {
	Status string `json:"status"`
}

type LeadOutput struct {
	Id                   string          `json:"id"`
	LandingPageId        string          `json:"landingpage_id"`
	LandingPageNome      string          `json:"landingpage_nome,omitempty"`
	Nome                 string          `json:"nome"`
	Email                string          `json:"email"`
	Telefone             string          `json:"telefone"`
	Whatsapp             string          `json:"whatsapp"`
	Mensagem             string          `json:"mensagem"`
	CamposPersonalizados json.RawMessage `json:"campos_personalizados,omitempty"`
	Origem               string          `json:"origem"`
	Status               string          `json:"status"`
	CriadoEm             time.Time       `json:"criado_em"`
}

type ListaLeadsOutput struct {
	Dados     []*LeadOutput `json:"dados"`
	Total     int           `json:"total"`
	Pagina    int           `json:"pagina"`
	PorPagina int           `json:"por_pagina"`
}
