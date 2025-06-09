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
	"github.com/MikhailRaia/url-shortener/internal/storage/postgres"
	"github.com/rs/zerolog/log"
)

type App struct {
	config    *config.Config
	handler   http.Handler
	dbStorage *postgres.Storage
}

func NewApp(cfg *config.Config) *App {
	logger.InitLogger()

	var urlStorage storage.URLStorage
	var dbStorage *postgres.Storage
	var err error

	// Если указан DSN для базы данных, пытаемся подключиться к PostgreSQL
	if cfg.DatabaseDSN != "" {
		dbStorage, err = postgres.NewStorage(cfg.DatabaseDSN)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize PostgreSQL storage")
		} else {
			log.Info().Msg("Using PostgreSQL storage")
			urlStorage = dbStorage
		}
	}

	// Если не удалось подключиться к PostgreSQL или DSN не указан, используем файловое хранилище
	if urlStorage == nil && cfg.FileStoragePath != "" {
		urlStorage, err = file.NewStorage(cfg.FileStoragePath)
		if err != nil {
			log.Error().Err(err).Str("path", cfg.FileStoragePath).Msg("Failed to initialize file storage, falling back to memory storage")
			urlStorage = memory.NewStorage()
		} else {
			log.Info().Str("path", cfg.FileStoragePath).Msg("Using file storage")
		}
	}

	// Если не удалось подключиться к PostgreSQL и файловое хранилище не указано или не удалось инициализировать,
	// используем хранилище в памяти
	if urlStorage == nil {
		urlStorage = memory.NewStorage()
		log.Info().Msg("Using memory storage")
	}

	urlService := service.NewURLService(urlStorage, cfg.BaseURL)

	// Передаем dbStorage как реализацию DBPinger, может быть nil, если PostgreSQL не используется
	httpHandler := handler.NewHandler(urlService, dbStorage)

	return &App{
		config:    cfg,
		handler:   httpHandler.RegisterRoutes(),
		dbStorage: dbStorage,
	}
}

func (a *App) Run() error {
	log.Info().Str("url", a.config.BaseURL).Str("address", a.config.ServerAddress).Msg("Starting server")

	// Закрываем соединение с базой данных при завершении работы
	if a.dbStorage != nil {
		defer a.dbStorage.Close()
	}

	return http.ListenAndServe(a.config.ServerAddress, a.handler)
}
