package app

import (
	"bytes"
	"github.com/MikhailRaia/url-shortener/internal/config"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApp_Integration(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080",
	}

	app := NewApp(cfg)

	server := httptest.NewServer(app.handler)
	defer server.Close()

	originalURL := "https://example.com"
	resp, err := http.Post(
		server.URL+"/",
		"text/plain",
		bytes.NewBufferString(originalURL),
	)

	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	shortenedURL := string(body)

	if !strings.HasPrefix(shortenedURL, cfg.BaseURL) {
		t.Errorf("Shortened URL %s does not start with base URL %s", shortenedURL, cfg.BaseURL)
	}

	id := strings.TrimPrefix(shortenedURL, cfg.BaseURL+"/")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err = client.Get(server.URL + "/" + id)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, got %d", http.StatusTemporaryRedirect, resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != originalURL {
		t.Errorf("Expected Location header %s, got %s", originalURL, location)
	}
}
