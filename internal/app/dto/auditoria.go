package dto

import (
	"encoding/json"
	"time"
)

type AuditoriaOutput struct {
	Id              string          `json:"id"`
	UsuarioId       *string         `json:"usuario_id,omitempty"`
	Acao            string          `json:"acao"`
	Entidade        string          `json:"entidade,omitempty"`
	EntidadeId      *string         `json:"entidade_id,omitempty"`
	DadosAnteriores json.RawMessage `json:"dados_anteriores,omitempty"`
	DadosNovos      json.RawMessage `json:"dados_novos,omitempty"`
	IP              string          `json:"ip"`
	UserAgent       string          `json:"user_agent,omitempty"`
	StatusCode      int             `json:"status_code"`
	TempoRespostaMs int             `json:"tempo_resposta_ms"`
	RequestId       string          `json:"request_id"`
	CriadoEm        time.Time       `json:"criado_em"`
}

type ListaAuditoriaOutput struct {
	Dados     []*AuditoriaOutput `json:"dados"`
	Total     int                `json:"total"`
	Pagina    int                `json:"pagina"`
	PorPagina int                `json:"por_pagina"`
}
