package dto

import "time"

type CriarUsuarioInput struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
	Role  string `json:"role,omitempty"`
}

type AtualizarUsuarioInput struct {
	Nome    string `json:"nome,omitempty"`
	Email   string `json:"email,omitempty"`
	Senha   string `json:"senha,omitempty"`
	Version int    `json:"version"`
}

type UsuarioOutput struct {
	Id           string    `json:"id"`
	Nome         string    `json:"nome"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
	Version      int       `json:"version"`
}

type ListaUsuariosOutput struct {
	Dados     []*UsuarioOutput `json:"dados"`
	Total     int              `json:"total"`
	Pagina    int              `json:"pagina"`
	PorPagina int              `json:"por_pagina"`
}
