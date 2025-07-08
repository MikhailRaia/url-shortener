package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MikhailRaia/url-shortener/internal/auth"
	"github.com/MikhailRaia/url-shortener/internal/config"
	"github.com/MikhailRaia/url-shortener/internal/handler"
	"github.com/MikhailRaia/url-shortener/internal/logger"
	"github.com/MikhailRaia/url-shortener/internal/middleware"
	"github.com/MikhailRaia/url-shortener/internal/service"
	"github.com/MikhailRaia/url-shortener/internal/storage"
	"github.com/MikhailRaia/url-shortener/internal/storage/file"
	"github.com/MikhailRaia/url-shortener/internal/storage/memory"
	"github.com/MikhailRaia/url-shortener/internal/storage/postgres"
	"github.com/MikhailRaia/url-shortener/internal/worker"
	"github.com/rs/zerolog/log"
)

type App struct {
	config         *config.Config
	handler        http.Handler
	dbStorage      *postgres.Storage
	jwtService     *auth.JWTService
	authMiddleware *middleware.AuthMiddleware
	deleteWorker   *worker.DeleteWorkerPool
}

func NewApp(cfg *config.Config) *App {
	logger.InitLogger()

	var urlStorage storage.URLStorage
	var dbStorage *postgres.Storage
	var err error

	if cfg.DatabaseDSN != "" {
		dbStorage, err = postgres.NewStorage(cfg.DatabaseDSN)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize PostgreSQL storage")
		} else {
			log.Info().Msg("Using PostgreSQL storage")
			urlStorage = dbStorage
		}
	}

	if urlStorage == nil && cfg.FileStoragePath != "" {
		urlStorage, err = file.NewStorage(cfg.FileStoragePath)
		if err != nil {
			log.Error().Err(err).Str("path", cfg.FileStoragePath).Msg("Failed to initialize file storage, falling back to memory storage")
			urlStorage = memory.NewStorage()
		} else {
			log.Info().Str("path", cfg.FileStoragePath).Msg("Using file storage")
		}
	}

	if urlStorage == nil {
		urlStorage = memory.NewStorage()
		log.Info().Msg("Using memory storage")
	}

	urlService := service.NewURLService(urlStorage, cfg.BaseURL)

	// Создаем JWT сервис
	jwtService := auth.NewJWTService(cfg.JWTSecretKey)

	// Создаем middleware для аутентификации
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	deleteWorkerConfig := worker.DefaultConfig()
	deleteWorker := worker.NewDeleteWorkerPool(urlService, deleteWorkerConfig)
	deleteWorker.Start()
	log.Info().Msg("Delete worker pool started")

	httpHandler := handler.NewHandlerWithDeleteWorker(urlService, dbStorage, deleteWorker)

	return &App{
		config:       cfg,
		handler:      httpHandler.RegisterRoutesWithAuth(authMiddleware),
		dbStorage:    dbStorage,
		jwtService:   jwtService,
		deleteWorker: deleteWorker,
	}
}

func (a *App) Run() error {
	log.Info().Str("url", a.config.BaseURL).Str("address", a.config.ServerAddress).Msg("Starting server")

	defer func() {
		if a.dbStorage != nil {
			log.Info().Msg("Closing database connection")
			a.dbStorage.Close()
		}

		if a.deleteWorker != nil {
			log.Info().Msg("Shutting down delete worker pool")
			if err := a.deleteWorker.Shutdown(10 * time.Second); err != nil {
				log.Error().Err(err).Msg("Error during worker pool shutdown")
			}
		}
	}()

	if err := http.ListenAndServe(a.config.ServerAddress, a.handler); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}
