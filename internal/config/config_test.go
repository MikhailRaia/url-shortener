package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.ServerAddress != ":8080" {
		t.Errorf("NewConfig() ServerAddress = %v, want %v", cfg.ServerAddress, ":8080")
	}

	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("NewConfig() BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:8080")
	}
}
