package dto

import "time"

type TemplateOutput struct {
	Id           string    `json:"id"`
	NegocioId    string    `json:"negocio_id,omitempty"`
	CategoriaId  string    `json:"categoria_id"`
	Tipo         string    `json:"tipo"`
	Origem       string    `json:"origem"`
	Nome         string    `json:"nome"`
	ArquivoURL   string    `json:"arquivo_url"`
	Conteudo     string    `json:"conteudo"`
	TemArquivo   bool      `json:"tem_arquivo"`
	NomeArquivo  string    `json:"nome_arquivo,omitempty"`
	ContentType  string    `json:"content_type,omitempty"`
	TamanhoBytes int64     `json:"tamanho_bytes,omitempty"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type CriarTemplateInput struct {
	NegocioId    string
	CategoriaId  string
	Tipo         string
	Origem       string
	Nome         string
	ArquivoURL   string
	Conteudo     string
	Arquivo      []byte
	NomeArquivo  string
	ContentType  string
	TamanhoBytes int64
}

type ListaTemplatesOutput struct {
	Dados     []*TemplateOutput `json:"dados"`
	Total     int               `json:"total"`
	Pagina    int               `json:"pagina"`
	PorPagina int               `json:"por_pagina"`
}
