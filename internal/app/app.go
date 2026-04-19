package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MikhailRaia/url-shortener/internal/auth"
	"github.com/MikhailRaia/url-shortener/internal/config"
	internalgrpc "github.com/MikhailRaia/url-shortener/internal/grpc"
	"github.com/MikhailRaia/url-shortener/internal/grpc/proto"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// App wires storage, services, middleware, and HTTP handlers and controls the server lifecycle.
type App struct {
	config         *config.Config
	handler        http.Handler
	dbStorage      *postgres.Storage
	jwtService     *auth.JWTService
	authMiddleware *middleware.AuthMiddleware
	deleteWorker   *worker.DeleteWorkerPool
	grpcServer     *grpc.Server
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

	httpHandler := handler.NewHandlerWithTrustedSubnet(urlService, dbStorage, deleteWorker, cfg.TrustedSubnet)

	// Инициализируем gRPC сервер
	var grpcOpts []grpc.ServerOption
	authInterceptor := internalgrpc.NewAuthInterceptor(jwtService)
	grpcOpts = append(grpcOpts, grpc.UnaryInterceptor(authInterceptor.Unary()))

	if cfg.EnableHTTPS {
		// Если сертификатов нет, они создадутся в Run() перед запуском серверов
		creds, err := credentials.NewServerTLSFromFile(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			log.Error().Err(err).Msg("gRPC: Failed to load TLS credentials")
		} else {
			grpcOpts = append(grpcOpts, grpc.Creds(creds))
		}
	}

	grpcServer := grpc.NewServer(grpcOpts...)
	shortenerServer := internalgrpc.NewShortenerServer(urlService)
	proto.RegisterShortenerServiceServer(grpcServer, shortenerServer)

	return &App{
		config:       cfg,
		handler:      httpHandler.RegisterRoutesWithAuth(authMiddleware),
		dbStorage:    dbStorage,
		jwtService:   jwtService,
		deleteWorker: deleteWorker,
		grpcServer:   grpcServer,
	}
}

// Run starts the HTTP and gRPC servers and performs graceful shutdown of resources on exit.
func (a *App) Run() error {
	log.Info().
		Str("url", a.config.BaseURL).
		Str("http_address", a.config.ServerAddress).
		Str("grpc_address", a.config.GRPCAddress).
		Bool("https", a.config.EnableHTTPS).
		Msg("Starting servers")

	defer a.cleanup()

	httpServer := a.setupHTTPServer()
	serverError := make(chan error, 2)

	a.startHTTPServer(httpServer, serverError)
	a.startGRPCServer(serverError)

	return a.handleShutdown(httpServer, serverError)
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

	if a.grpcServer != nil {
		log.Info().Msg("Stopping gRPC server")
		a.grpcServer.GracefulStop()
	}
}

func (a *App) setupHTTPServer() *http.Server {
	return &http.Server{
		Addr:    a.config.ServerAddress,
		Handler: a.handler,
	}
}

func (a *App) startHTTPServer(server *http.Server, serverError chan<- error) {
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

func (a *App) startGRPCServer(serverError chan<- error) {
	go func() {
		listen, err := net.Listen("tcp", a.config.GRPCAddress)
		if err != nil {
			serverError <- fmt.Errorf("failed to listen for gRPC: %w", err)
			return
		}

		log.Info().Str("address", a.config.GRPCAddress).Msg("gRPC server listening")
		if err := a.grpcServer.Serve(listen); err != nil {
			serverError <- fmt.Errorf("failed to start gRPC server: %w", err)
		}
	}()
}

func (a *App) handleShutdown(httpServer *http.Server, serverError <-chan error) error {
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

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown failed: %w", err)
		}
	}

	return nil
}
