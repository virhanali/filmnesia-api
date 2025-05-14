package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/virhanali/filmnesia/config"
)

func NewPostgresConnection(cfg *config.DBConfig) *sqlx.DB {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
