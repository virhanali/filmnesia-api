package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/virhanali/filmnesia/user-service/internal/config"
	"github.com/virhanali/filmnesia/user-service/internal/platform/database"
	userHttp "github.com/virhanali/filmnesia/user-service/internal/user/delivery/http"
	userRepo "github.com/virhanali/filmnesia/user-service/internal/user/repository"
	userUsecase "github.com/virhanali/filmnesia/user-service/internal/user/usecase"
)

func main() {
	cfg, err := config.LoadConfig("../")
	if err != nil {
		log.Fatalf("FATAL: Cannot load config: %v", err)
	}

	db, err := database.NewPostgresSQLDB(cfg)
	if err != nil {
		log.Fatalf("FATAL: Cannot connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("WARNING: Cannot close database connection: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}(db)

	pgUserRepo := userRepo.NewPostgresUserRepository(db)

	ucase := userUsecase.NewUserUsecase(pgUserRepo)

	userHandler := userHttp.NewUserHandler(ucase)

	router := gin.Default()

	userHandler.RegisterRoutes(router)
	router.GET("/health", func(c *gin.Context) {
		errDB := db.PingContext(c.Request.Context())
		if errDB != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "unhealthy", "db_error": errDB.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	serverAddr := ":" + cfg.ServicePort
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}
	go func() {
		log.Printf("INFO: Server HTTP User Service started on port %s", cfg.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Cannot run HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("INFO: Server HTTP User Service sedang dimatikan...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Server HTTP User Service failed to shutdown gracefully: %v", err)
	}

	log.Println("INFO: Server HTTP User Service successfully shutdown.")
}
