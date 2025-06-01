// api-gateway/cmd/api/main.go
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

	"github.com/virhanali/filmnesia/api-gateway/internal/config"
	"github.com/virhanali/filmnesia/api-gateway/internal/router"
)

func main() {
	cfg, err := config.LoadConfig("../../")
	if err != nil {
		log.Fatalf("FATAL: Failed to load API Gateway configuration: %v", err)
	}

	if cfg.UserServiceURL == "" {
		log.Fatal("FATAL: USER_SERVICE_URL must be configured.")
	}

	r := router.SetupRouter(cfg)

	serverAddr := ":" + cfg.APIGatewayPort
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	go func() {
		log.Printf("INFO: API Gateway starting on port %s", cfg.APIGatewayPort)
		log.Printf("INFO: Proxying /api/v1/users/* to %s", cfg.UserServiceURL)
		if errSrv := srv.ListenAndServe(); errSrv != nil && !errors.Is(errSrv, http.ErrServerClosed) {
			log.Fatalf("FATAL: API Gateway ListenAndServe error: %v", errSrv)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("INFO: API Gateway shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if errShut := srv.Shutdown(ctx); errShut != nil {
		log.Fatalf("FATAL: API Gateway server failed to shutdown gracefully: %v", errShut)
	}

	log.Println("INFO: API Gateway server shutdown complete.")
}
