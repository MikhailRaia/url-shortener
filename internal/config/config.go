package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds application configuration loaded from flags and environment variables.
// All fields can be overridden by environment variables with the prefix pattern
// (e.g., SERVER_ADDRESS, BASE_URL, DATABASE_DSN).
type Config struct {
	// ServerAddress is the TCP address the server listens on (flag: -a, default: :8080)
	ServerAddress string
	// BaseURL is the base URL for shortened URLs (flag: -b, default: http://localhost:8080)
	BaseURL string
	// FileStoragePath is the path to file-based storage (flag: -f, default: ~/.url-shortener/storage.json)
	FileStoragePath string
	// DatabaseDSN is the PostgreSQL connection string (flag: -d, optional)
	DatabaseDSN string
	// JWTSecretKey is the secret key for signing JWT tokens (flag: -jwt)
	JWTSecretKey string
	// EnableHTTPS indicates if the server should use HTTPS (flag: -s)
	EnableHTTPS bool
	// MaxProcs is the GOMAXPROCS value (flag: -p, 0=auto)
	MaxProcs int
}

// NewConfig returns a Config initialized from command-line flags and environment variables.
func NewConfig() *Config {
	cfg := &Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: getDefaultStoragePath(),
		DatabaseDSN:     "",
		JWTSecretKey:    "default-secret-key-change-in-production",
		EnableHTTPS:     false,
		MaxProcs:        0,
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address (e.g. localhost:8888)")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL for shortened URLs (e.g. http://localhost:8000)")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database connection string (e.g. postgres://username:password@localhost:5432/database_name)")
	flag.StringVar(&cfg.JWTSecretKey, "jwt", cfg.JWTSecretKey, "JWT secret key for signing tokens")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS")
	flag.IntVar(&cfg.MaxProcs, "p", cfg.MaxProcs, "GOMAXPROCS value (0=auto)")

	flag.Parse()

	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		if b, err := strconv.ParseBool(envEnableHTTPS); err == nil {
			cfg.EnableHTTPS = b
		}
	}

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
