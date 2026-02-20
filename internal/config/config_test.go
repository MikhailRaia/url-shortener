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

func TestNewConfigTrustedSubnetFlag(t *testing.T) {
	oldArgs := os.Args
	oldTrustedSubnet := os.Getenv("TRUSTED_SUBNET")

	defer func() {
		os.Args = oldArgs
		os.Setenv("TRUSTED_SUBNET", oldTrustedSubnet)
	}()

	os.Unsetenv("TRUSTED_SUBNET")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-t", "192.168.1.0/24"}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.TrustedSubnet != "192.168.1.0/24" {
		t.Errorf("NewConfig() TrustedSubnet = %v, want %v", cfg.TrustedSubnet, "192.168.1.0/24")
	}
}

func TestNewConfigTrustedSubnetEnv(t *testing.T) {
	oldArgs := os.Args
	oldTrustedSubnet := os.Getenv("TRUSTED_SUBNET")

	defer func() {
		os.Args = oldArgs
		os.Setenv("TRUSTED_SUBNET", oldTrustedSubnet)
	}()

	os.Setenv("TRUSTED_SUBNET", "10.0.0.0/8")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.TrustedSubnet != "10.0.0.0/8" {
		t.Errorf("NewConfig() TrustedSubnet = %v, want %v", cfg.TrustedSubnet, "10.0.0.0/8")
	}
}

func TestNewConfigTrustedSubnetEnvOverridesFlag(t *testing.T) {
	oldArgs := os.Args
	oldTrustedSubnet := os.Getenv("TRUSTED_SUBNET")

	defer func() {
		os.Args = oldArgs
		os.Setenv("TRUSTED_SUBNET", oldTrustedSubnet)
	}()

	os.Setenv("TRUSTED_SUBNET", "172.16.0.0/12")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-t", "192.168.1.0/24"}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.TrustedSubnet != "172.16.0.0/12" {
		t.Errorf("NewConfig() TrustedSubnet = %v, want %v", cfg.TrustedSubnet, "172.16.0.0/12")
	}
}

func TestNewConfigTrustedSubnetDefault(t *testing.T) {
	oldArgs := os.Args
	oldTrustedSubnet := os.Getenv("TRUSTED_SUBNET")

	defer func() {
		os.Args = oldArgs
		os.Setenv("TRUSTED_SUBNET", oldTrustedSubnet)
	}()

	os.Unsetenv("TRUSTED_SUBNET")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.TrustedSubnet != "" {
		t.Errorf("NewConfig() TrustedSubnet = %v, want %v", cfg.TrustedSubnet, "")
	}
}
