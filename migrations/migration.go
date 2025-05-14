package migrations

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

type Migration struct {
	Name  string
	Query string
}

func RunMigrations(db *sqlx.DB, migrations []Migration) {
	for _, m := range migrations {
		log.Printf("Running migration: %s", m.Name)
		_, err := db.Exec(m.Query)
		if err != nil {
			log.Fatalf("Migration %s failed: %v", m.Name, err)
		}
	}
	log.Println("All migrations applied successfully.")
}

func RunSQLMigrationsFromDir(db *sqlx.DB, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			path := filepath.Join(dir, file.Name())
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			log.Printf("Running migration: %s", file.Name())
			_, err = db.Exec(string(content))
			if err != nil {
				return err
			}
		}
	}
	log.Println("All migrations applied successfully.")
	return nil
}
