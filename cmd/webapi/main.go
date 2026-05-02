package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"buildnix-backend/cmd/internal/bootstrap"
	"buildnix-backend/internal/interface/webapi"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	container, err := bootstrap.New(ctx)
	if err != nil {
		log.Fatalf("inicializar dependencias: %v", err)
	}
	defer container.Close()

	servidor, err := (&webapi.Server{
		AuthUseCase:        container.AuthUseCase,
		UsuarioUseCase:     container.UsuarioUseCase,
		ClienteUseCase:     container.ClienteUseCase,
		DashboardUseCase:   container.DashboardUseCase,
		NegocioUseCase:     container.NegocioUseCase,
		CategoriaUseCase:   container.CategoriaUseCase,
		TemplateUseCase:    container.TemplateUseCase,
		LandingPageUseCase: container.LandingPageUseCase,
		LeadUseCase:        container.LeadUseCase,
		ExportacaoUseCase:  container.ExportacaoUseCase,
		AuditoriaRepo:      container.AuditoriaRepository,
		JWTServico:         container.JWTServico,
		GoogleClientID:     container.Config.GoogleClientID,
		GoogleClientSecret: container.Config.GoogleClientSecret,
		GoogleRedirectURL:  container.Config.GoogleRedirectURL,
	}).Start(context.Background(), ":"+container.Config.ServerPort)
	if err != nil {
		log.Fatalf("iniciar webapi: %v", err)
	}
	log.Printf("servidor iniciado na porta %s", container.Config.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("encerrando servidor...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := servidor.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("encerramento forcado: %v", err)
	}
	log.Println("servidor encerrado")
}
