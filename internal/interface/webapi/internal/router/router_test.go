package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServirOpenAPI(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	response := httptest.NewRecorder()

	servirOpenAPI(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status esperado %d, recebido %d", http.StatusOK, response.Code)
	}
	if !strings.Contains(response.Body.String(), "openapi: 3.0.3") {
		t.Fatal("resposta nao contem especificacao OpenAPI")
	}
}

func TestOpenAPIContemRotasDoRouter(t *testing.T) {
	conteudo := string(openapiYAML)
	rotas := map[string][]string{
		"/swagger/openapi.yaml":       {"get"},
		"/auth/login":                 {"post"},
		"/auth/refresh":               {"post"},
		"/auth/logout":                {"post"},
		"/auth/google/login":          {"get"},
		"/auth/google/callback":       {"get"},
		"/usuarios":                   {"post", "get"},
		"/usuarios/me":                {"get"},
		"/usuarios/{id}":              {"put", "delete"},
		"/clientes":                   {"get", "post"},
		"/clientes/me":                {"get"},
		"/clientes/{id}":              {"put", "delete"},
		"/usuarios-clientes":          {"post"},
		"/dashboard":                  {"get"},
		"/negocios":                   {"get", "post"},
		"/negocios/{id}":              {"put", "delete"},
		"/categorias":                 {"get", "post"},
		"/categorias/{id}":            {"put", "delete"},
		"/templates":                  {"get", "post"},
		"/templates/{id}":             {"delete"},
		"/landingpages":               {"get", "post"},
		"/landingpages/{id}":          {"get", "put", "delete"},
		"/landingpages/{id}/autosave": {"patch"},
		"/landingpages/{id}/lock":     {"patch"},
		"/landingpages/{id}/unlock":   {"patch"},
		"/landingpages/{id}/exportar": {"post"},
		"/leads":                      {"post", "get"},
		"/leads/{id}":                 {"get"},
		"/leads/{id}/status":          {"patch"},
		"/exportacoes":                {"get"},
		"/exportacoes/{id}/download":  {"get"},
		"/auditoria":                  {"get"},
		"/auditoria/{id}":             {"get"},
	}

	for rota, metodos := range rotas {
		bloco := blocoRotaOpenAPI(conteudo, rota)
		if bloco == "" {
			t.Fatalf("openapi nao contem rota %s", rota)
		}
		for _, metodo := range metodos {
			if !strings.Contains(bloco, "\n    "+metodo+":") {
				t.Fatalf("openapi nao contem metodo %s para rota %s", metodo, rota)
			}
		}
	}
}

func blocoRotaOpenAPI(conteudo, rota string) string {
	inicio := strings.Index(conteudo, "  "+rota+":")
	if inicio == -1 {
		return ""
	}
	resto := conteudo[inicio+1:]
	fim := strings.Index(resto, "\n  /")
	if fim == -1 {
		return resto
	}
	return resto[:fim]
}
