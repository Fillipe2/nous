package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"buildnix-backend/internal/app/dto"
	ucauth "buildnix-backend/internal/app/usecase/auth"
	"buildnix-backend/internal/domain/usuario"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const cookieOAuthState = "oauth_state"

type AuthHandler struct {
	useCase      *ucauth.UseCase
	googleConfig *oauth2.Config
}

func NewAuthHandler(useCase *ucauth.UseCase, googleClientID, googleClientSecret, googleRedirectURL string) *AuthHandler {
	googleConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &AuthHandler{useCase: useCase, googleConfig: googleConfig}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.Login(r.Context(), input)
	if errors.Is(err, usuario.ErrNaoAutorizado) || errors.Is(err, usuario.ErrNaoEncontrado) {
		responderErro(w, http.StatusUnauthorized, "credenciais invalidas")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var input dto.RefreshTokenInput
	if err := decodificarJSON(r, &input); err != nil {
		responderErro(w, http.StatusBadRequest, "payload invalido")
		return
	}

	output, err := h.useCase.RefreshToken(r.Context(), input)
	if errors.Is(err, usuario.ErrNaoAutorizado) {
		responderErro(w, http.StatusUnauthorized, "token invalido ou expirado")
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if !h.googleConfigurado() {
		responderErro(w, http.StatusServiceUnavailable, "google oauth nao configurado")
		return
	}

	nonce := gerarStateOAuth()
	http.SetCookie(w, &http.Cookie{
		Name:     cookieOAuthState,
		Value:    nonce,
		MaxAge:   300,
		HttpOnly: true,
		Secure:   h.googleCallbackSeguro(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	url := h.googleConfig.AuthCodeURL(nonce, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func gerarStateOAuth() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !h.googleConfigurado() {
		responderErro(w, http.StatusServiceUnavailable, "google oauth nao configurado")
		return
	}
	if erroOAuth := r.URL.Query().Get("error"); erroOAuth != "" {
		responderErro(w, http.StatusUnauthorized, "falha na autenticacao google")
		return
	}

	// Validar CSRF state
	cookieState, err := r.Cookie(cookieOAuthState)
	if err != nil || cookieState.Value == "" {
		responderErro(w, http.StatusBadRequest, "state oauth ausente")
		return
	}
	if r.URL.Query().Get("state") != cookieState.Value {
		responderErro(w, http.StatusBadRequest, "state oauth invalido")
		return
	}
	// Invalidar o cookie após uso
	http.SetCookie(w, &http.Cookie{
		Name:     cookieOAuthState,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.googleCallbackSeguro(),
		Path:     "/",
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		responderErro(w, http.StatusBadRequest, "codigo ausente")
		return
	}

	token, err := h.googleConfig.Exchange(r.Context(), code)
	if err != nil {
		responderErro(w, http.StatusUnauthorized, "falha na autenticacao google")
		return
	}

	userInfo, err := obterInfoGoogleUsuario(r.Context(), h.googleConfig, token)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro ao obter dados do google")
		return
	}

	output, err := h.useCase.LoginGoogle(r.Context(), ucauth.GoogleCallbackInput{
		ProviderId:    userInfo.Sub,
		ProviderEmail: userInfo.Email,
		Nome:          userInfo.Name,
	})
	if err != nil {
		responderErro(w, http.StatusInternalServerError, "erro ao processar login google")
		return
	}

	responderJSON(w, http.StatusOK, output)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var input dto.LogoutInput
	if r.Body != nil {
		_ = decodificarJSON(r, &input)
	}
	if err := h.useCase.Logout(r.Context(), input); err != nil {
		responderErro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	responderJSON(w, http.StatusOK, map[string]string{"mensagem": "logout realizado com sucesso"})
}

func (h *AuthHandler) googleConfigurado() bool {
	return strings.TrimSpace(h.googleConfig.ClientID) != "" &&
		strings.TrimSpace(h.googleConfig.ClientSecret) != "" &&
		strings.TrimSpace(h.googleConfig.RedirectURL) != ""
}

func (h *AuthHandler) googleCallbackSeguro() bool {
	callbackURL, err := url.Parse(h.googleConfig.RedirectURL)
	return err == nil && callbackURL.Scheme == "https"
}
