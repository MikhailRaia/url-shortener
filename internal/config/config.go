package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds application configuration loaded from flags and environment variables.
// All fields can be overridden by environment variables with the prefix pattern
// (e.g., SERVER_ADDRESS, BASE_URL, DATABASE_DSN).
type Config struct {
	// ServerAddress is the TCP address the server listens on (flag: -a, default: :8080)
	ServerAddress string `json:"server_address"`
	// BaseURL is the base URL for shortened URLs (flag: -b, default: http://localhost:8080)
	BaseURL string `json:"base_url"`
	// FileStoragePath is the path to file-based storage (flag: -f, default: ~/.url-shortener/storage.json)
	FileStoragePath string `json:"file_storage_path"`
	// DatabaseDSN is the PostgreSQL connection string (flag: -d, optional)
	DatabaseDSN string `json:"database_dsn"`
	// JWTSecretKey is the secret key for signing JWT tokens (flag: -jwt)
	JWTSecretKey string `json:"jwt_secret_key"`
	// EnableHTTPS indicates if the server should use HTTPS (flag: -s)
	EnableHTTPS bool `json:"enable_https"`
	// MaxProcs is the GOMAXPROCS value (flag: -p, 0=auto)
	MaxProcs int `json:"max_procs"`
	// ConfigPath is the path to the JSON configuration file (flag: -c, -config)
	ConfigPath string
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

	// 1. Define all flags
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address (e.g. localhost:8888)")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL for shortened URLs (e.g. http://localhost:8000)")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database connection string (e.g. postgres://username:password@localhost:5432/database_name)")
	flag.StringVar(&cfg.JWTSecretKey, "jwt", cfg.JWTSecretKey, "JWT secret key for signing tokens")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS")
	flag.IntVar(&cfg.MaxProcs, "p", cfg.MaxProcs, "GOMAXPROCS value (0=auto)")
	flag.StringVar(&cfg.ConfigPath, "c", "", "Path to JSON configuration file")
	flag.StringVar(&cfg.ConfigPath, "config", "", "Path to JSON configuration file (long form)")

	// 2. Determine config path from env or command line before full flag.Parse()
	configPath := os.Getenv("CONFIG")
	for i, arg := range os.Args {
		if (arg == "-c" || arg == "-config" || arg == "--config") && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
			break
		}
		if strings.HasPrefix(arg, "-c=") {
			configPath = strings.TrimPrefix(arg, "-c=")
			break
		}
		if strings.HasPrefix(arg, "-config=") {
			configPath = strings.TrimPrefix(arg, "-config=")
			break
		}
		if strings.HasPrefix(arg, "--config=") {
			configPath = strings.TrimPrefix(arg, "--config=")
			break
		}
	}

	// 3. Load from JSON if path is provided
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			var jsonCfg Config
			if err := json.Unmarshal(data, &jsonCfg); err == nil {
				// Update defaults with JSON values
				if jsonCfg.ServerAddress != "" {
					cfg.ServerAddress = jsonCfg.ServerAddress
				}
				if jsonCfg.BaseURL != "" {
					cfg.BaseURL = jsonCfg.BaseURL
				}
				if jsonCfg.FileStoragePath != "" {
					cfg.FileStoragePath = jsonCfg.FileStoragePath
				}
				if jsonCfg.DatabaseDSN != "" {
					cfg.DatabaseDSN = jsonCfg.DatabaseDSN
				}
				if jsonCfg.JWTSecretKey != "" {
					cfg.JWTSecretKey = jsonCfg.JWTSecretKey
				}
				if jsonCfg.EnableHTTPS {
					cfg.EnableHTTPS = jsonCfg.EnableHTTPS
				}
				if jsonCfg.MaxProcs != 0 {
					cfg.MaxProcs = jsonCfg.MaxProcs
				}
			}
		}
	}

	// 4. Parse flags (will overwrite JSON values if flag is provided)
	flag.Parse()

	// 5. Apply environment variables (highest priority)
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
