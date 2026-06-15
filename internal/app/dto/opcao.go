package dto

type OpcaoOutput struct {
	Id          string `json:"id"`
	Nome        string `json:"nome"`
	Label       string `json:"label"`
	Status      string `json:"status,omitempty"`
	Tipo        string `json:"tipo,omitempty"`
	Origem      string `json:"origem,omitempty"`
	NegocioId   string `json:"negocio_id,omitempty"`
	CategoriaId string `json:"categoria_id,omitempty"`
}

type ListaOpcoesOutput struct {
	Dados     []*OpcaoOutput `json:"dados"`
	Total     int            `json:"total"`
	Pagina    int            `json:"pagina"`
	PorPagina int            `json:"por_pagina"`
}
