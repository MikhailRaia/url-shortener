package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Hello, Yandex!"}`))
	})

	gzipHandler := GzipMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding to be gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzipped response: %v", err)
	}

	expected := `{"message":"Hello, Yandex!"}`
	if string(body) != expected {
		t.Errorf("Expected response body to be %s, got %s", expected, string(body))
	}
}

func TestGzipMiddleware_NoGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Hello, Yandex!"}`))
	})

	gzipHandler := GzipMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Errorf("Expected Content-Encoding not to be gzip")
	}

	body := rec.Body.String()
	expected := `{"message":"Hello, Yandex!"}`
	if body != expected {
		t.Errorf("Expected response body to be %s, got %s", expected, body)
	}
}

func TestGzipReader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	gzipHandler := GzipReader(handler)

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write([]byte("Hello, Yandex!"))
	if err != nil {
		t.Fatalf("Failed to write to gzip writer: %v", err)
	}
	gzWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	expected := "Hello, Yandex!"
	if body != expected {
		t.Errorf("Expected response body to be %s, got %s", expected, body)
	}
}

func TestGzipReader_NoGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	gzipHandler := GzipReader(handler)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("Hello, Yandex!"))

	rec := httptest.NewRecorder()

	gzipHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	expected := "Hello, Yandex!"
	if body != expected {
		t.Errorf("Expected response body to be %s, got %s", expected, body)
	}
}
