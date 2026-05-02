package handler

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestValidarZIPAceitaArquivoSeguro(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.Create("index.html")
	if err != nil {
		t.Fatalf("criar entrada zip: %v", err)
	}
	if _, err := file.Write([]byte("ok")); err != nil {
		t.Fatalf("escrever entrada zip: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("fechar zip: %v", err)
	}

	if err := validarZIP(buffer.Bytes()); err != nil {
		t.Fatalf("zip seguro rejeitado: %v", err)
	}
}

func TestValidarZIPRejeitaPathTraversal(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	if _, err := writer.Create("../secrets.env"); err != nil {
		t.Fatalf("criar entrada zip: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("fechar zip: %v", err)
	}

	if err := validarZIP(buffer.Bytes()); err == nil {
		t.Fatal("zip com path traversal deveria ser rejeitado")
	}
}

func TestContentTypeZIP(t *testing.T) {
	aceitos := []string{"", "application/zip", "application/x-zip-compressed", "application/octet-stream; charset=binary"}
	for _, contentType := range aceitos {
		if !contentTypeZIP(contentType) {
			t.Fatalf("content type deveria ser aceito: %s", contentType)
		}
	}

	if contentTypeZIP(strings.TrimSpace("text/plain")) {
		t.Fatal("content type text/plain nao deveria ser aceito")
	}
}
