package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGoogleLoginRequerConfiguracao(t *testing.T) {
	handler := NewAuthHandler(nil, "", "", "")
	request := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	response := httptest.NewRecorder()

	handler.GoogleLogin(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status esperado %d, recebido %d", http.StatusServiceUnavailable, response.Code)
	}
}

func TestGoogleLoginRedirecionaParaGoogle(t *testing.T) {
	handler := NewAuthHandler(nil, "cliente-test", "segredo-test", "http://localhost:8080/auth/google/callback")
	request := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	response := httptest.NewRecorder()

	handler.GoogleLogin(response, request)

	if response.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status esperado %d, recebido %d", http.StatusTemporaryRedirect, response.Code)
	}

	location := response.Header().Get("Location")
	parsedLocation, err := url.Parse(location)
	if err != nil {
		t.Fatalf("location invalido: %v", err)
	}
	if !strings.Contains(parsedLocation.Host, "google") {
		t.Fatalf("host esperado do google, recebido %s", parsedLocation.Host)
	}
	if parsedLocation.Query().Get("client_id") != "cliente-test" {
		t.Fatalf("client_id nao enviado para google")
	}
	if parsedLocation.Query().Get("state") == "" {
		t.Fatalf("state oauth nao enviado")
	}

	cookies := response.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != cookieOAuthState || cookies[0].Value == "" {
		t.Fatalf("cookie oauth_state nao configurado")
	}
}
