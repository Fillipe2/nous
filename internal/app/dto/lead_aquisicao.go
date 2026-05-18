package dto

import (
	"encoding/json"
	"time"
)

type RegistrarLeadAquisicaoInput struct {
	Nome                     string          `json:"nome"`
	Email                    string          `json:"email"`
	Telefone                 string          `json:"telefone,omitempty"`
	CategoriaPretendidaId    string          `json:"categoria_pretendida_id,omitempty"`
	CategoriaPretendidaLivre string          `json:"categoria_pretendida_livre,omitempty"`
	Cidade                   string          `json:"cidade,omitempty"`
	Estado                   string          `json:"estado,omitempty"`
	Porte                    string          `json:"porte,omitempty"`
	Interesses               json.RawMessage `json:"interesses,omitempty"`
	Mensagem                 string          `json:"mensagem,omitempty"`
	Origem                   string          `json:"origem,omitempty"`
	UTM                      *UTMInput       `json:"utm,omitempty"`
	// Honeypot: campo escondido no formulario; bots tendem a preencher.
	Honeypot string `json:"website,omitempty"`
}

type AtualizarFunilLeadAquisicaoInput struct {
	StatusFunil           string `json:"status_funil"`
	ConvertidoEmUsuarioId string `json:"convertido_em_usuario_id,omitempty"`
}

type LeadAquisicaoOutput struct {
	Id                       string          `json:"id"`
	Nome                     string          `json:"nome"`
	Email                    string          `json:"email"`
	Telefone                 string          `json:"telefone,omitempty"`
	CategoriaPretendidaId    string          `json:"categoria_pretendida_id,omitempty"`
	CategoriaPretendidaLivre string          `json:"categoria_pretendida_livre,omitempty"`
	Cidade                   string          `json:"cidade,omitempty"`
	Estado                   string          `json:"estado,omitempty"`
	Porte                    string          `json:"porte,omitempty"`
	Interesses               json.RawMessage `json:"interesses,omitempty"`
	Mensagem                 string          `json:"mensagem,omitempty"`
	Origem                   string          `json:"origem,omitempty"`
	UTM                      *UTMInput       `json:"utm,omitempty"`
	StatusFunil              string          `json:"status_funil"`
	ConvertidoEmUsuarioId    string          `json:"convertido_em_usuario_id,omitempty"`
	CriadoEm                 time.Time       `json:"criado_em"`
	AtualizadoEm             time.Time       `json:"atualizado_em"`
}

type ListaLeadsAquisicaoOutput struct {
	Dados     []*LeadAquisicaoOutput `json:"dados"`
	Total     int                    `json:"total"`
	Pagina    int                    `json:"pagina"`
	PorPagina int                    `json:"por_pagina"`
}
