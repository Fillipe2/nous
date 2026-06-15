package dto

import "time"

type CriarUsuarioInput struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type AtualizarUsuarioInput struct {
	Nome    string `json:"nome,omitempty"`
	Email   string `json:"email,omitempty"`
	Version int    `json:"version"`
}

type AlterarSenhaInput struct {
	SenhaAtual string `json:"senha_atual"`
	NovaSenha  string `json:"nova_senha"`
}

type AlterarRoleInput struct {
	Role string `json:"role"`
}

type AlterarStatusInput struct {
	Status string `json:"status"`
}

type UsuarioOutput struct {
	Id            string     `json:"id"`
	Nome          string     `json:"nome"`
	Email         string     `json:"email"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	EmailVerified bool       `json:"email_verified"`
	UltimoLoginEm *time.Time `json:"ultimo_login_em,omitempty"`
	CriadoEm      time.Time  `json:"criado_em"`
	AtualizadoEm  time.Time  `json:"atualizado_em"`
	Version       int        `json:"version"`
}

type ListaUsuariosOutput struct {
	Dados     []*UsuarioOutput `json:"dados"`
	Total     int              `json:"total"`
	Pagina    int              `json:"pagina"`
	PorPagina int              `json:"por_pagina"`
}
