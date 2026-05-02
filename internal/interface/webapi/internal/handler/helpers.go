package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func responderJSON(w http.ResponseWriter, status int, dados interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dados)
}

func responderErro(w http.ResponseWriter, status int, mensagem string) {
	responderJSON(w, status, map[string]string{
		"erro":       mensagem,
		"codigo":     codigoErro(status),
		"mensagem":   mensagem,
		"request_id": requestIDHeader(w),
	})
}

func responderErroRequest(w http.ResponseWriter, r *http.Request, status int, mensagem string) {
	responderJSON(w, status, map[string]string{
		"erro":       mensagem,
		"codigo":     codigoErro(status),
		"mensagem":   mensagem,
		"request_id": chimiddleware.GetReqID(r.Context()),
	})
}

func requestIDHeader(w http.ResponseWriter) string {
	if requestID := w.Header().Get("X-Request-Id"); requestID != "" {
		return requestID
	}
	return w.Header().Get("X-Request-ID")
}

func codigoErro(status int) string {
	codigo := strings.ToLower(http.StatusText(status))
	codigo = strings.ReplaceAll(codigo, " ", "_")
	if codigo == "" {
		return "erro"
	}
	return codigo
}

func decodificarJSON(r *http.Request, destino interface{}) error {
	return json.NewDecoder(r.Body).Decode(destino)
}

func paginacao(r *http.Request) (pagina, porPagina int) {
	pagina = 1
	porPagina = 20

	if p := r.URL.Query().Get("pagina"); p != "" {
		if v := atoiSafe(p); v > 0 {
			pagina = v
		}
	}
	if pp := r.URL.Query().Get("por_pagina"); pp != "" {
		if v := atoiSafe(pp); v > 0 && v <= 100 {
			porPagina = v
		}
	}
	return pagina, porPagina
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
