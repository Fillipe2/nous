package webapi

import (
	"context"
	"testing"
)

func TestServerStartValidaJWTServico(t *testing.T) {
	server := &Server{}

	if _, err := server.Start(context.Background(), ":0"); err == nil {
		t.Fatal("erro esperado quando JWTServico nao esta configurado")
	}
}
