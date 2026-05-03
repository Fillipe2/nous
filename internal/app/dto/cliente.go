package dto

import "time"

type CriarClienteInput struct {
	Nome     string `json:"nome"`
	Email    string `json:"email,omitempty"`
	Telefone string `json:"telefone,omitempty"`
	Whatsapp string `json:"whatsapp,omitempty"`
}

type AtualizarClienteInput struct {
	Nome     string `json:"nome,omitempty"`
	Email    string `json:"email,omitempty"`
	Telefone string `json:"telefone,omitempty"`
	Whatsapp string `json:"whatsapp,omitempty"`
	Version  int    `json:"version"`
}

type ClienteOutput struct {
	Id           string    `json:"id"`
	Nome         string    `json:"nome"`
	Email        string    `json:"email"`
	Telefone     string    `json:"telefone,omitempty"`
	Whatsapp     string    `json:"whatsapp,omitempty"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
	Version      int       `json:"version"`
}

type ListaClientesOutput struct {
	Dados     []*ClienteOutput `json:"dados"`
	Total     int              `json:"total"`
	Pagina    int              `json:"pagina"`
	PorPagina int              `json:"por_pagina"`
}

type ClienteBasicoOutput struct {
	Id       string `json:"id"`
	Nome     string `json:"nome"`
	Email    string `json:"email,omitempty"`
	Telefone string `json:"telefone,omitempty"`
	Whatsapp string `json:"whatsapp,omitempty"`
	Status   string `json:"status"`
}

type ListaClientesBasicosOutput struct {
	Dados     []*ClienteBasicoOutput `json:"dados"`
	Total     int                    `json:"total"`
	Pagina    int                    `json:"pagina"`
	PorPagina int                    `json:"por_pagina"`
}

type VincularUsuarioClienteInput struct {
	UsuarioId string `json:"usuario_id"`
	ClienteId string `json:"cliente_id"`
}
