package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	infrajwt "nous/internal/infra/jwt"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const claimsKey contextKey = "claims"

type Claims struct {
	UsuarioId string
	Email     string
	Role      string
}

func Autenticar(jwtServico *infrajwt.Servico) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				responderErro(w, r, http.StatusUnauthorized, "token nao fornecido")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				responderErro(w, r, http.StatusUnauthorized, "formato de token invalido")
				return
			}

			jwtClaims, err := jwtServico.ValidarToken(parts[1])
			if err != nil {
				responderErro(w, r, http.StatusUnauthorized, "token invalido ou expirado")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, &Claims{
				UsuarioId: jwtClaims.UsuarioId,
				Email:     jwtClaims.Email,
				Role:      jwtClaims.Role,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ApenasAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ObterClaims(r.Context())
		if claims == nil || claims.Role != "admin" {
			responderErro(w, r, http.StatusForbidden, "acesso nao autorizado")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func responderErro(w http.ResponseWriter, r *http.Request, status int, mensagem string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"erro":       mensagem,
		"codigo":     codigoErro(status),
		"mensagem":   mensagem,
		"request_id": chimiddleware.GetReqID(r.Context()),
	})
}

func codigoErro(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusTooManyRequests:
		return "too_many_requests"
	default:
		return "erro"
	}
}

func ObterClaims(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}
