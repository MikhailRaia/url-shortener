package main

import (
	"github.com/MikhailRaia/url-shortener/internal/app"
	"github.com/MikhailRaia/url-shortener/internal/config"
	"log"
)

func main() {
	cfg := config.NewConfig()

	application := app.NewApp(cfg)
	if err := application.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
