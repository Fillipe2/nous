package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"buildnix-backend/internal/domain/auditoria"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func RegistrarAuditoria(repository auditoria.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inicio := time.Now()
			response := &auditResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(response, r)

			claims := ObterClaims(r.Context())
			if claims == nil {
				return
			}

			acao := r.Method + " " + r.URL.Path
			ip := ClienteIP(r)

			a := &auditoria.Auditoria{
				UsuarioId:       &claims.UsuarioId,
				Acao:            acao,
				IP:              ip,
				UserAgent:       r.UserAgent(),
				StatusCode:      response.statusCode,
				TempoRespostaMs: int(time.Since(inicio).Milliseconds()),
				RequestId:       chimiddleware.GetReqID(r.Context()),
				CriadoEm:        time.Now(),
			}

			dados, _ := json.Marshal(map[string]string{
				"metodo":     r.Method,
				"path":       r.URL.Path,
				"request_id": a.RequestId,
			})
			a.DadosNovos = dados

			if err := repository.Salvar(r.Context(), a); err != nil {
				log.Printf("salvar auditoria: %v", err)
			}
		})
	}
}
