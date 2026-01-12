package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds application configuration loaded from flags and environment variables.
type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	JWTSecretKey    string
	MaxProcs        int
}

// NewConfig returns a Config initialized from command-line flags and environment variables.
func NewConfig() *Config {
	cfg := &Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: getDefaultStoragePath(),
		DatabaseDSN:     "",
		JWTSecretKey:    "default-secret-key-change-in-production",
		MaxProcs:        0,
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address (e.g. localhost:8888)")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL for shortened URLs (e.g. http://localhost:8000)")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database connection string (e.g. postgres://username:password@localhost:5432/database_name)")
	flag.StringVar(&cfg.JWTSecretKey, "s", cfg.JWTSecretKey, "JWT secret key for signing tokens")
	flag.IntVar(&cfg.MaxProcs, "p", cfg.MaxProcs, "GOMAXPROCS value (0=auto)")

	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

	if envJWTSecretKey := os.Getenv("JWT_SECRET_KEY"); envJWTSecretKey != "" {
		cfg.JWTSecretKey = envJWTSecretKey
	}

	if envMaxProcs := os.Getenv("MAX_PROCS"); envMaxProcs != "" {
		if n, err := strconv.Atoi(envMaxProcs); err == nil {
			cfg.MaxProcs = n
		}
	}

	return cfg
}

func getDefaultStoragePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "storage.json"
	}
	return filepath.Join(homeDir, ".url-shortener", "storage.json")
}
