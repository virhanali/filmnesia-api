package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/virhanali/filmnesia/user-service/internal/config"
	"github.com/virhanali/filmnesia/user-service/internal/platform/database"
	"github.com/virhanali/filmnesia/user-service/internal/platform/messagebroker"
	userHttp "github.com/virhanali/filmnesia/user-service/internal/user/delivery/http"
	userRepo "github.com/virhanali/filmnesia/user-service/internal/user/repository"
	userUsecase "github.com/virhanali/filmnesia/user-service/internal/user/usecase"
)

type App struct {
	Config    config.Config
	DB        *sql.DB
	Router    *gin.Engine
	Publisher *messagebroker.RabbitMQPublisher
}

func NewApp(configPath string) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.JWTSecretKey == "" {
		return nil, errors.New("JWT_SECRET_KEY must not be empty in configuration")
	}
	if cfg.JWTExpirationHours <= 0 {
		log.Printf("WARNING: JWT_EXPIRATION_HOURS is invalid (%d), using default 24 hours.", cfg.JWTExpirationHours)
		cfg.JWTExpirationHours = 24
	}

	if cfg.RabbitMQURL == "" {
		log.Println("WARNING: RabbitMQURL is empty in configuration. Skipping RabbitMQ publisher initialization.")
	}

	db, err := database.NewPostgresSQLDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	publisher, err := messagebroker.NewRabbitMQPublisher(cfg)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize RabbitMQ publisher: %w", err)
	}

	pgUserRepo := userRepo.NewPostgresUserRepository(db)
	ucase := userUsecase.NewUserUsecase(pgUserRepo, cfg, publisher)
	userHandler := userHttp.NewUserHandler(ucase)

	router := gin.Default()

	authMiddleware := userHttp.AuthMiddleware(cfg.JWTSecretKey)

	publicRoutes := router.Group("/api/v1/users")
	{
		publicRoutes.POST("/register", userHandler.Register)
		publicRoutes.POST("/login", userHandler.Login)
		publicRoutes.GET("/email/:email", userHandler.GetUserByEmail)
		publicRoutes.GET("/username/:username", userHandler.GetUserByUsername)
	}

	authenticatedRoutes := router.Group("/api/v1")
	authenticatedRoutes.Use(authMiddleware)
	{
		authenticatedRoutes.GET("/users/me", userHandler.GetMyProfile)
		authenticatedRoutes.GET("/users/:id", userHandler.GetUserByID)
		authenticatedRoutes.PUT("/users/:id", userHandler.UpdateUser)
		authenticatedRoutes.DELETE("/users/:id", userHandler.DeleteUser)
	}

	router.GET("/health", func(c *gin.Context) {
		errDB := db.PingContext(c.Request.Context())
		if errDB != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "unhealthy", "db_error": errDB.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	return &App{
		Config: cfg,
		DB:     db,
		Router: router,
	}, nil
}

func (a *App) Run() {

	if a.Publisher != nil {
		defer a.Publisher.Close()
	}
	if a.DB != nil {
		defer func() {
			if err := a.DB.Close(); err != nil {
				log.Printf("WARNING: Failed to close database connection: %v", err)
			} else {
				log.Println("Database connection closed.")
			}
		}()
	}

	serverAddr := ":" + a.Config.ServicePort
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: a.Router,
	}

	go func() {
		log.Printf("INFO: User Service HTTP server starting on port %s", a.Config.ServicePort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("FATAL: Failed to run HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("INFO: User Service HTTP server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: User Service HTTP server failed to shutdown gracefully: %v", err)
	}

	log.Println("INFO: User Service HTTP server shutdown complete.")
}
