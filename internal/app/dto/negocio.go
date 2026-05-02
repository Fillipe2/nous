package dto

import "time"

type CriarNegocioInput struct {
	ClienteId string `json:"cliente_id"`
	Nome      string `json:"nome"`
}

type AtualizarNegocioInput struct {
	Nome    string `json:"nome,omitempty"`
	Version int    `json:"version"`
}

type NegocioOutput struct {
	Id           string    `json:"id"`
	ClienteId    string    `json:"cliente_id"`
	Nome         string    `json:"nome"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
	Version      int       `json:"version"`
}

type ListaNegociosOutput struct {
	Dados     []*NegocioOutput `json:"dados"`
	Total     int              `json:"total"`
	Pagina    int              `json:"pagina"`
	PorPagina int              `json:"por_pagina"`
}
