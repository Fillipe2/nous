package dto

import "time"

type CriarUnidadeInput struct {
	NegocioId string `json:"negocio_id"`
	Nome      string `json:"nome"`
}

type AtualizarUnidadeInput struct {
	Nome    string `json:"nome,omitempty"`
	Version int    `json:"version"`
}

type UnidadeOutput struct {
	Id           string    `json:"id"`
	NegocioId    string    `json:"negocio_id"`
	Nome         string    `json:"nome"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
	Version      int       `json:"version"`
}

type ListaUnidadesOutput struct {
	Dados     []*UnidadeOutput `json:"dados"`
	Total     int              `json:"total"`
	Pagina    int              `json:"pagina"`
	PorPagina int              `json:"por_pagina"`
}
