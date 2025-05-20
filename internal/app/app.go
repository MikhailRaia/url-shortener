package app

import (
	"net/http"

	"github.com/MikhailRaia/url-shortener/internal/config"
	"github.com/MikhailRaia/url-shortener/internal/handler"
	"github.com/MikhailRaia/url-shortener/internal/logger"
	"github.com/MikhailRaia/url-shortener/internal/service"
	"github.com/MikhailRaia/url-shortener/internal/storage"
	"github.com/MikhailRaia/url-shortener/internal/storage/file"
	"github.com/MikhailRaia/url-shortener/internal/storage/memory"
	"github.com/rs/zerolog/log"
)

type App struct {
	config  *config.Config
	handler http.Handler
}

func NewApp(cfg *config.Config) *App {
	logger.InitLogger()

	var urlStorage storage.URLStorage
	var err error

	if cfg.FileStoragePath != "" {
		urlStorage, err = file.NewStorage(cfg.FileStoragePath)
		if err != nil {
			log.Error().Err(err).Str("path", cfg.FileStoragePath).Msg("Failed to initialize file storage, falling back to memory storage")
			urlStorage = memory.NewStorage()
		} else {
			log.Info().Str("path", cfg.FileStoragePath).Msg("Using file storage")
		}
	} else {
		urlStorage = memory.NewStorage()
		log.Info().Msg("Using memory storage")
	}

	urlService := service.NewURLService(urlStorage, cfg.BaseURL)

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
