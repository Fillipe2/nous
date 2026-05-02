package handler

import (
	"net/http"

	"buildnix-backend/internal/domain/auditoria"

	"github.com/go-chi/chi/v5"
)

type AuditoriaHandler struct {
	repository auditoria.Repository
}

func NewAuditoriaHandler(repository auditoria.Repository) *AuditoriaHandler {
	return &AuditoriaHandler{repository: repository}
}

func (h *AuditoriaHandler) ListarAuditoria(w http.ResponseWriter, r *http.Request) {
	pagina, porPagina := paginacao(r)
	filtro := auditoria.FiltroListagem{
		UsuarioId:  r.URL.Query().Get("usuario_id"),
		Acao:       r.URL.Query().Get("acao"),
		Entidade:   r.URL.Query().Get("entidade"),
		EntidadeId: r.URL.Query().Get("entidade_id"),
		StatusCode: r.URL.Query().Get("status_code"),
		RequestId:  r.URL.Query().Get("request_id"),
		DataInicio: r.URL.Query().Get("data_inicio"),
		DataFim:    r.URL.Query().Get("data_fim"),
		Pagina:     pagina,
		PorPagina:  porPagina,
		OrdenarPor: r.URL.Query().Get("ordenar_por"),
		Ordem:      r.URL.Query().Get("ordem"),
	}

	auditorias, total, err := h.repository.Listar(r.Context(), filtro)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, map[string]interface{}{
		"dados":      auditorias,
		"total":      total,
		"pagina":     pagina,
		"por_pagina": porPagina,
	})
}

func (h *AuditoriaHandler) ObterAuditoria(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	a, err := h.repository.BuscarPorID(r.Context(), id)
	if err != nil {
		responderErro(w, http.StatusNotFound, "registro nao encontrado")
		return
	}
	responderJSON(w, http.StatusOK, a)
}
