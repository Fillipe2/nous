package dto

import "time"

type CriarCategoriaInput struct {
	Nome string `json:"nome"`
}

type AtualizarCategoriaInput struct {
	Nome    string `json:"nome,omitempty"`
	Version int    `json:"version"`
}

type CategoriaOutput struct {
	Id           string    `json:"id"`
	Nome         string    `json:"nome"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
	Version      int       `json:"version"`
}

type ListaCategoriasOutput struct {
	Dados     []*CategoriaOutput `json:"dados"`
	Total     int                `json:"total"`
	Pagina    int                `json:"pagina"`
	PorPagina int                `json:"por_pagina"`
}
