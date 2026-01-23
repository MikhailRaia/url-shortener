package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigWithJSON(t *testing.T) {
	oldArgs := os.Args
	oldServerAddress := os.Getenv("SERVER_ADDRESS")
	oldBaseURL := os.Getenv("BASE_URL")
	oldConfig := os.Getenv("CONFIG")

	defer func() {
		os.Args = oldArgs
		os.Setenv("SERVER_ADDRESS", oldServerAddress)
		os.Setenv("BASE_URL", oldBaseURL)
		os.Setenv("CONFIG", oldConfig)
	}()

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("CONFIG")

	// Create temp JSON config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	jsonContent := `{
		"server_address": "json:8080",
		"base_url": "http://json",
		"enable_https": true
	}`
	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-c", configPath}

	cfg := NewConfig()

	if cfg.ServerAddress != "json:8080" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, "json:8080")
	}

	if cfg.BaseURL != "http://json" {
		t.Errorf("NewConfig() BaseURL = %v, want %v", cfg.BaseURL, "http://json")
	}

	if !cfg.EnableHTTPS {
		t.Errorf("NewConfig() EnableHTTPS = %v, want %v", cfg.EnableHTTPS, true)
	}
}

func TestNewConfigJSONPriority(t *testing.T) {
	oldArgs := os.Args
	oldServerAddress := os.Getenv("SERVER_ADDRESS")
	oldBaseURL := os.Getenv("BASE_URL")
	oldConfig := os.Getenv("CONFIG")

	defer func() {
		os.Args = oldArgs
		os.Setenv("SERVER_ADDRESS", oldServerAddress)
		os.Setenv("BASE_URL", oldBaseURL)
		os.Setenv("CONFIG", oldConfig)
	}()

	// 1. JSON says "json:8080"
	// 2. Flag says "flag:8080"
	// 3. Env says "env:8080"
	// Env should win.

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	jsonContent := `{"server_address": "json:8080"}`
	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("SERVER_ADDRESS", "env:8080")
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-c", configPath, "-a", "flag:8080"}

	cfg := NewConfig()

	if cfg.ServerAddress != "env:8080" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, "env:8080")
	}
}
