package dto

import "time"

type CriarNegocioInput struct {
	Nome      string `json:"nome"`
	Email     string `json:"email,omitempty"`
	Telefone  string `json:"telefone,omitempty"`
	Whatsapp  string `json:"whatsapp,omitempty"`
	Documento string `json:"documento,omitempty"`
}

type AtualizarNegocioInput struct {
	Nome      string `json:"nome,omitempty"`
	Email     string `json:"email,omitempty"`
	Telefone  string `json:"telefone,omitempty"`
	Whatsapp  string `json:"whatsapp,omitempty"`
	Documento string `json:"documento,omitempty"`
	Version   int    `json:"version"`
}

type NegocioOutput struct {
	Id           string    `json:"id"`
	Nome         string    `json:"nome"`
	Email        string    `json:"email"`
	Telefone     string    `json:"telefone,omitempty"`
	Whatsapp     string    `json:"whatsapp,omitempty"`
	Documento    string    `json:"documento,omitempty"`
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

type NegocioBasicoOutput struct {
	Id        string `json:"id"`
	Nome      string `json:"nome"`
	Email     string `json:"email,omitempty"`
	Telefone  string `json:"telefone,omitempty"`
	Whatsapp  string `json:"whatsapp,omitempty"`
	Documento string `json:"documento,omitempty"`
	Status    string `json:"status"`
}

type ListaNegociosBasicosOutput struct {
	Dados     []*NegocioBasicoOutput `json:"dados"`
	Total     int                    `json:"total"`
	Pagina    int                    `json:"pagina"`
	PorPagina int                    `json:"por_pagina"`
}

type VincularUsuarioNegocioInput struct {
	UsuarioId string `json:"usuario_id"`
	NegocioId string `json:"negocio_id"`
}
