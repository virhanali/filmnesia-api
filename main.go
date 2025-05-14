package main

import (
	"fmt"

	"github.com/virhanali/filmnesia/config"
	"github.com/virhanali/filmnesia/migrations"
)

func main() {
	dbConfig := config.LoadDBConfig()
	db := migrations.NewPostgresConnection(dbConfig)
	defer db.Close()

	if db == nil {
		fmt.Println("Koneksi ke Postgres gagal!")
		return
	}
	fmt.Println("Koneksi ke Postgres berhasil!")

	err := migrations.RunSQLMigrationsFromDir(db, "migrations/sql")
	if err != nil {
		fmt.Println("Error migrasi:", err)
	} else {
		fmt.Println("Migrasi sukses!")
	}
}
