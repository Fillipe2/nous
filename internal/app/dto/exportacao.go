package dto

import "time"

type ExportacaoOutput struct {
	Id            string    `json:"id"`
	LandingPageId string    `json:"landingpage_id"`
	UsuarioId     string    `json:"usuario_id"`
	NomeArquivo   string    `json:"nome_arquivo"`
	ContentType   string    `json:"content_type"`
	TamanhoBytes  int64     `json:"tamanho_bytes"`
	Hash          string    `json:"hash"`
	CriadoEm      time.Time `json:"criado_em"`
}

type ListaExportacoesOutput struct {
	Dados     []*ExportacaoOutput `json:"dados"`
	Total     int                 `json:"total"`
	Pagina    int                 `json:"pagina"`
	PorPagina int                 `json:"por_pagina"`
}
