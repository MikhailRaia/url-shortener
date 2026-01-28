package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
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

	server := &http.Server{
		Addr:    a.config.ServerAddress,
		Handler: a.handler,
	}

	serverError := make(chan error, 1)

	go func() {
		if a.config.EnableHTTPS {
			certFile := "cert.pem"
			keyFile := "key.pem"

			if _, err := os.Stat(certFile); os.IsNotExist(err) {
				if err := createSelfSignedCert(certFile, keyFile); err != nil {
					serverError <- fmt.Errorf("failed to create self-signed certificate: %w", err)
					return
				}
				log.Info().Msg("Self-signed certificate created")
			}

			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				serverError <- fmt.Errorf("failed to start HTTPS server: %w", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverError <- fmt.Errorf("failed to start HTTP server: %w", err)
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case err := <-serverError:
		return err
	case sig := <-stop:
		log.Info().Str("signal", sig.String()).Msg("Shutting down gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}

	return nil
}

func createSelfSignedCert(certFile, keyFile string) error {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2025),
		Subject: pkix.Name{
			Organization: []string{"URL Shortener"},
			Country:      []string{"RU"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	pub := &priv.PublicKey
	certBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return err
	}
	if err := certOut.Close(); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return err
	}
	if err := keyOut.Close(); err != nil {
		return err
	}

	return nil
}
