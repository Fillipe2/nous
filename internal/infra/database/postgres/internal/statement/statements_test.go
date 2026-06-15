package statement

import "testing"

func TestLoadStatements(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	loaded := Get()
	if loaded.AtualizarNegocio == "" {
		t.Fatal("AtualizarNegocio nao foi carregado")
	}
	if loaded.AtualizarLandingPage == "" {
		t.Fatal("AtualizarLandingPage nao foi carregado")
	}
	if loaded.SalvarAuditoria == "" {
		t.Fatal("SalvarAuditoria nao foi carregado")
	}
}
