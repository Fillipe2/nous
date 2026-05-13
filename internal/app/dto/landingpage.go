package dto

import (
	"encoding/json"
	"time"
)

type CriarLandingPageInput struct {
	ClienteId   string `json:"cliente_id"`
	NegocioId   string `json:"negocio_id,omitempty"`
	CategoriaId string `json:"categoria_id,omitempty"`
	TemplateId  string `json:"template_id,omitempty"`
	Tipo        string `json:"tipo,omitempty"`
	Nome        string `json:"nome"`
	Slug        string `json:"slug,omitempty"`
	Conteudo    string `json:"conteudo,omitempty"`
}

type AtualizarLandingPageInput struct {
	NegocioId   string `json:"negocio_id,omitempty"`
	CategoriaId string `json:"categoria_id,omitempty"`
	TemplateId  string `json:"template_id,omitempty"`
	Tipo        string `json:"tipo,omitempty"`
	Nome        string `json:"nome,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Conteudo    string `json:"conteudo,omitempty"`
	Status      string `json:"status,omitempty"`
	Version     int    `json:"version"`
}

type AutosaveInput struct {
	AutosaveData json.RawMessage `json:"autosave_data"`
	Version      int             `json:"version"`
}

type AutosaveOutput struct {
	Mensagem string `json:"mensagem"`
	Version  int    `json:"version"`
}

type LandingPageOutput struct {
	Id             string          `json:"id"`
	ClienteId      string          `json:"cliente_id"`
	NegocioId      string          `json:"negocio_id"`
	CategoriaId    string          `json:"categoria_id"`
	TemplateId     string          `json:"template_id"`
	Tipo           string          `json:"tipo"`
	Nome           string          `json:"nome"`
	Slug           string          `json:"slug,omitempty"`
	Conteudo       string          `json:"conteudo"`
	Status         string          `json:"status"`
	IsDraft        bool            `json:"is_draft"`
	AutosaveData   json.RawMessage `json:"autosave_data,omitempty"`
	LastAutosaveEm *time.Time      `json:"last_autosave_em,omitempty"`
	LockedBy       *string         `json:"locked_by,omitempty"`
	LockedAt       *time.Time      `json:"locked_at,omitempty"`
	UrlPublica     *string         `json:"url_publica,omitempty"`
	DeployId       *string         `json:"deploy_id,omitempty"`
	PublicadaEm    *time.Time      `json:"publicada_em,omitempty"`
	CriadoEm       time.Time       `json:"criado_em"`
	AtualizadoEm   time.Time       `json:"atualizado_em"`
	Version        int             `json:"version"`
}

type ListaLandingPagesOutput struct {
	Dados     []*LandingPageOutput `json:"dados"`
	Total     int                  `json:"total"`
	Pagina    int                  `json:"pagina"`
	PorPagina int                  `json:"por_pagina"`
}

type PublicarLandingPageOutput struct {
	Mensagem   string `json:"mensagem"`
	UrlPublica string `json:"url_publica"`
	DeployId   string `json:"deploy_id"`
}
