package app

import (
	"log"
	"net/http"

	"github.com/MikhailRaia/url-shortener/internal/config"
	"github.com/MikhailRaia/url-shortener/internal/handler"
	"github.com/MikhailRaia/url-shortener/internal/service"
	"github.com/MikhailRaia/url-shortener/internal/storage/memory"
)

type App struct {
	config  *config.Config
	handler http.Handler
}

func NewApp(cfg *config.Config) *App {
	storage := memory.NewStorage()

	urlService := service.NewURLService(storage, cfg.BaseURL)

	httpHandler := handler.NewHandler(urlService)

	return &App{
		config:  cfg,
		handler: httpHandler.RegisterRoutes(),
	}
}

func (a *App) Run() error {
	log.Printf("Starting server at %s", a.config.BaseURL)
	return http.ListenAndServe(a.config.ServerAddress, a.handler)
}
