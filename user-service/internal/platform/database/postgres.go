package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/virhanali/filmnesia/user-service/internal/config"
)

func NewPostgresSQLDB(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	log.Printf("Trying to connect to database: %s", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to database")

	return db, nil
}
