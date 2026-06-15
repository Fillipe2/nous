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
	t.Setenv("OAUTH_STATE_SAMESITE", "")
	handler := NewAuthHandler(nil, "negocio-test", "segredo-test", "http://localhost:8080/auth/google/callback")
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
	if parsedLocation.Query().Get("client_id") != "negocio-test" {
		t.Fatalf("client_id nao enviado para google")
	}
	if parsedLocation.Query().Get("state") == "" {
		t.Fatalf("state oauth nao enviado")
	}

	cookies := response.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != cookieOAuthState || cookies[0].Value == "" {
		t.Fatalf("cookie oauth_state nao configurado")
	}
	if cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite esperado Lax, recebido %v", cookies[0].SameSite)
	}
}

func TestGoogleLoginNaoUsaSameSiteNoneSemHTTPS(t *testing.T) {
	t.Setenv("OAUTH_STATE_SAMESITE", "none")
	handler := NewAuthHandler(nil, "negocio-test", "segredo-test", "http://localhost:8080/auth/google/callback")
	request := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	response := httptest.NewRecorder()

	handler.GoogleLogin(response, request)

	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("cookie oauth_state nao configurado")
	}
	if cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite none sem HTTPS deveria cair para Lax, recebido %v", cookies[0].SameSite)
	}
	if cookies[0].Secure {
		t.Fatal("cookie nao deveria ser Secure com callback HTTP")
	}
}

func TestSameSiteOAuthStateConfiguravel(t *testing.T) {
	t.Setenv("OAUTH_STATE_SAMESITE", "lax")
	if sameSiteOAuthState(false) != http.SameSiteLaxMode {
		t.Fatal("SameSite lax nao aplicado")
	}

	t.Setenv("OAUTH_STATE_SAMESITE", "strict")
	if sameSiteOAuthState(false) != http.SameSiteStrictMode {
		t.Fatal("SameSite strict nao aplicado")
	}

	t.Setenv("OAUTH_STATE_SAMESITE", "none")
	if sameSiteOAuthState(true) != http.SameSiteNoneMode {
		t.Fatal("SameSite none nao aplicado")
	}
	if sameSiteOAuthState(false) != http.SameSiteLaxMode {
		t.Fatal("SameSite none sem HTTPS deveria cair para lax")
	}

	t.Setenv("OAUTH_STATE_SAMESITE", "valor-invalido")
	if sameSiteOAuthState(false) != http.SameSiteLaxMode {
		t.Fatal("SameSite invalido deveria cair para lax")
	}
}
