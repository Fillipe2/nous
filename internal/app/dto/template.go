package dto

import "time"

type TemplateOutput struct {
	Id           string    `json:"id"`
	ClienteId    string    `json:"cliente_id,omitempty"`
	CategoriaId  string    `json:"categoria_id"`
	Tipo         string    `json:"tipo"`
	Origem       string    `json:"origem"`
	Nome         string    `json:"nome"`
	ArquivoURL   string    `json:"arquivo_url"`
	Conteudo     string    `json:"conteudo"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type CriarTemplateInput struct {
	ClienteId   string
	CategoriaId string
	Tipo        string
	Origem      string
	Nome        string
	ArquivoURL  string
	Conteudo    string
}

type ListaTemplatesOutput struct {
	Dados     []*TemplateOutput `json:"dados"`
	Total     int               `json:"total"`
	Pagina    int               `json:"pagina"`
	PorPagina int               `json:"por_pagina"`
}
