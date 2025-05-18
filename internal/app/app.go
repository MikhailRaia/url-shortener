package app

import (
	"net/http"

	"github.com/MikhailRaia/url-shortener/internal/config"
	"github.com/MikhailRaia/url-shortener/internal/handler"
	"github.com/MikhailRaia/url-shortener/internal/logger"
	"github.com/MikhailRaia/url-shortener/internal/service"
	"github.com/MikhailRaia/url-shortener/internal/storage/memory"
	"github.com/rs/zerolog/log"
)

type App struct {
	config  *config.Config
	handler http.Handler
}

func NewApp(cfg *config.Config) *App {
	logger.InitLogger()

	storage := memory.NewStorage()

	urlService := service.NewURLService(storage, cfg.BaseURL)

	httpHandler := handler.NewHandler(urlService)

	return &App{
		config:  cfg,
		handler: httpHandler.RegisterRoutes(),
	}
}

func (a *App) Run() error {
	log.Info().Str("url", a.config.BaseURL).Str("address", a.config.ServerAddress).Msg("Starting server")
	return http.ListenAndServe(a.config.ServerAddress, a.handler)
}
