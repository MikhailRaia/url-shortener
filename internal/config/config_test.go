package config

import (
	"flag"
	"os"
	"testing"
)

func TestNewConfigDefault(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"cmd"}

	cfg := NewConfig()

	if cfg.ServerAddress != ":8080" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, ":8080")
	}

	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("NewConfig() BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:8080")
	}
}

func TestNewConfigWithArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"cmd", "-a", "localhost:8888", "-b", "http://localhost:8000"}

	cfg := NewConfig()

	if cfg.ServerAddress != "localhost:8888" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, "localhost:8888")
	}

	if cfg.BaseURL != "http://localhost:8000" {
		t.Errorf("NewConfig() BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:8000")
	}
}
