package main

import (
	"log"

	"github.com/virhanali/filmnesia/user-service/internal/app"
)

func main() {
	application, err := app.NewApp("../")
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize application: %v", err)
	}

	application.Run()
}
