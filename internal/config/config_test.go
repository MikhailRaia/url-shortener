package config

import (
	"flag"
	"os"
	"testing"
)

func TestNewConfigDefault(t *testing.T) {
	oldArgs := os.Args
	oldServerAddress := os.Getenv("SERVER_ADDRESS")
	oldBaseURL := os.Getenv("BASE_URL")

	defer func() {
		os.Args = oldArgs
		os.Setenv("SERVER_ADDRESS", oldServerAddress)
		os.Setenv("BASE_URL", oldBaseURL)
	}()

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.ServerAddress != ":8080" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, ":8080")
	}

	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("NewConfig() BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:8080")
	}
}

func TestNewConfigWithArgs(t *testing.T) {
	oldArgs := os.Args
	oldServerAddress := os.Getenv("SERVER_ADDRESS")
	oldBaseURL := os.Getenv("BASE_URL")

	defer func() {
		os.Args = oldArgs
		os.Setenv("SERVER_ADDRESS", oldServerAddress)
		os.Setenv("BASE_URL", oldBaseURL)
	}()

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-a", "localhost:8888", "-b", "http://localhost:8000"}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.ServerAddress != "localhost:8888" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, "localhost:8888")
	}

	if cfg.BaseURL != "http://localhost:8000" {
		t.Errorf("NewConfig() BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:8000")
	}
}

func TestNewConfigWithEnv(t *testing.T) {
	oldArgs := os.Args
	oldServerAddress := os.Getenv("SERVER_ADDRESS")
	oldBaseURL := os.Getenv("BASE_URL")

	defer func() {
		os.Args = oldArgs
		os.Setenv("SERVER_ADDRESS", oldServerAddress)
		os.Setenv("BASE_URL", oldBaseURL)
	}()

	os.Setenv("SERVER_ADDRESS", "localhost:9999")
	os.Setenv("BASE_URL", "http://localhost:9000")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.ServerAddress != "localhost:9999" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, "localhost:9999")
	}

	if cfg.BaseURL != "http://localhost:9000" {
		t.Errorf("NewConfig() BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:9000")
	}
}

func TestNewConfigEnvOverridesArgs(t *testing.T) {
	oldArgs := os.Args
	oldServerAddress := os.Getenv("SERVER_ADDRESS")
	oldBaseURL := os.Getenv("BASE_URL")

	defer func() {
		os.Args = oldArgs
		os.Setenv("SERVER_ADDRESS", oldServerAddress)
		os.Setenv("BASE_URL", oldBaseURL)
	}()

	os.Setenv("SERVER_ADDRESS", "localhost:9999")
	os.Setenv("BASE_URL", "http://localhost:9000")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-a", "localhost:8888", "-b", "http://localhost:8000"}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.ServerAddress != "localhost:9999" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, "localhost:9999")
	}

	if cfg.BaseURL != "http://localhost:9000" {
		t.Errorf("NewConfig() BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:9000")
	}
}
