package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"buildnix-backend/cmd/internal/bootstrap"
	"buildnix-backend/internal/interface/webapi"

	"github.com/joho/godotenv"
)

func main() {
	carregarDotenv()

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()

	container, err := bootstrap.New(startupCtx)
	if err != nil {
		log.Fatalf("inicializar dependencias: %v", err)
	}
	defer container.Close()

	runtimeCtx, runtimeCancel := context.WithCancel(context.Background())
	defer runtimeCancel()
	container.StartCleanupJobs(runtimeCtx)

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
		ProspectoUseCase:   container.ProspectoUseCase,
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
	runtimeCancel()
	webapi.CloseSharedResources()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := servidor.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("encerramento forcado: %v", err)
	}
	log.Println("servidor encerrado")
}

func carregarDotenv() {
	caminho := os.Getenv("DOTENV_PATH")
	if caminho == "" {
		caminho = ".env"
	}
	if _, err := os.Stat(caminho); err != nil {
		return
	}
	if err := godotenv.Load(caminho); err != nil {
		log.Printf("ignorando %s: %v", caminho, err)
		return
	}
	log.Printf("variaveis carregadas de %s", caminho)
}
