package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	"github.com/MikhailRaia/url-shortener/internal/tls"
	"github.com/MikhailRaia/url-shortener/internal/worker"
	"github.com/rs/zerolog/log"
)

// App wires storage, services, middleware, and HTTP handlers and controls the server lifecycle.
type App struct {
	config         *config.Config
	handler        http.Handler
	dbStorage      *postgres.Storage
	jwtService     *auth.JWTService
	authMiddleware *middleware.AuthMiddleware
	deleteWorker   *worker.DeleteWorkerPool
}

// NewApp creates and initializes application dependencies and HTTP routes.
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

// Run starts the HTTP server and performs graceful shutdown of resources on exit.
func (a *App) Run() error {
	log.Info().Str("url", a.config.BaseURL).Str("address", a.config.ServerAddress).Bool("https", a.config.EnableHTTPS).Msg("Starting server")

	defer a.cleanup()

	server := a.setupServer()
	serverError := make(chan error, 1)

	a.startServer(server, serverError)

	return a.handleShutdown(server, serverError)
}

func (a *App) cleanup() {
	if a.dbStorage != nil {
		log.Info().Msg("Closing database connection")
		a.dbStorage.Close()
	}

	if a.deleteWorker != nil {
		log.Info().Msg("Shutting down delete worker pool")
		timeout := time.Duration(a.config.WorkerShutdownTimeout) * time.Second
		if err := a.deleteWorker.Shutdown(timeout); err != nil {
			log.Error().Err(err).Msg("Error during worker pool shutdown")
		}
	}
}

func (a *App) setupServer() *http.Server {
	return &http.Server{
		Addr:    a.config.ServerAddress,
		Handler: a.handler,
	}
}

func (a *App) startServer(server *http.Server, serverError chan<- error) {
	go func() {
		if a.config.EnableHTTPS {
			if _, err := os.Stat(a.config.CertFile); os.IsNotExist(err) {
				if err := tls.CreateSelfSignedCert(a.config.CertFile, a.config.KeyFile); err != nil {
					serverError <- fmt.Errorf("failed to create self-signed certificate: %w", err)
					return
				}
				log.Info().Msg("Self-signed certificate created")
			}

			if err := server.ListenAndServeTLS(a.config.CertFile, a.config.KeyFile); err != nil && err != http.ErrServerClosed {
				serverError <- fmt.Errorf("failed to start HTTPS server: %w", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverError <- fmt.Errorf("failed to start HTTP server: %w", err)
			}
		}
	}()
}

func (a *App) handleShutdown(server *http.Server, serverError <-chan error) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case err := <-serverError:
		return err
	case sig := <-stop:
		log.Info().Str("signal", sig.String()).Msg("Shutting down gracefully...")

		timeout := time.Duration(a.config.ShutdownTimeout) * time.Second
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}

	return nil
}
