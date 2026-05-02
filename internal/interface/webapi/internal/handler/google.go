package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

type googleUsuarioInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

func obterInfoGoogleUsuario(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*googleUsuarioInfo, error) {
	client := config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("obter info google usuario request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("obter info google usuario status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("obter info google usuario ler body: %w", err)
	}

	var info googleUsuarioInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("obter info google usuario deserializar: %w", err)
	}
	if info.Sub == "" || info.Email == "" {
		return nil, fmt.Errorf("obter info google usuario incompleto")
	}
	if !info.EmailVerified {
		return nil, fmt.Errorf("email google nao verificado")
	}

	return &info, nil
}
