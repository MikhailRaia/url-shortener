package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockURLService struct {
	ShortenURLFunc     func(originalURL string) (string, error)
	GetOriginalURLFunc func(id string) (string, bool)
}

func (m *MockURLService) ShortenURL(originalURL string) (string, error) {
	return m.ShortenURLFunc(originalURL)
}

func (m *MockURLService) GetOriginalURL(id string) (string, bool) {
	return m.GetOriginalURLFunc(id)
}

func TestHandleShortenJSON(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		contentType        string
		mockShortenURLFunc func(originalURL string) (string, error)
		expectedStatus     int
		expectedResponse   *ShortenResponse
	}{
		{
			name: "Valid JSON request",
			requestBody: ShortenRequest{
				URL: "https://practicum.yandex.ru",
			},
			contentType: "application/json",
			mockShortenURLFunc: func(originalURL string) (string, error) {
				return "http://localhost:8080/abc123", nil
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: &ShortenResponse{
				Result: "http://localhost:8080/abc123",
			},
		},
		{
			name: "Empty URL in request",
			requestBody: ShortenRequest{
				URL: "",
			},
			contentType: "application/json",
			mockShortenURLFunc: func(originalURL string) (string, error) {
				return "", nil
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
		},
		{
			name:        "Invalid content type",
			requestBody: ShortenRequest{URL: "https://practicum.yandex.ru"},
			contentType: "text/plain",
			mockShortenURLFunc: func(originalURL string) (string, error) {
				return "", nil
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
		},
		{
			name:        "Invalid JSON format",
			requestBody: "not a json",
			contentType: "application/json",
			mockShortenURLFunc: func(originalURL string) (string, error) {
				return "", nil
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
		},
		{
			name: "Service error",
			requestBody: ShortenRequest{
				URL: "https://practicum.yandex.ru",
			},
			contentType: "application/json",
			mockShortenURLFunc: func(originalURL string) (string, error) {
				return "", errors.New("service error")
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody []byte
			var err error

			switch v := tt.requestBody.(type) {
			case string:
				requestBody = []byte(v)
			default:
				requestBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()

			mockService := &MockURLService{
				ShortenURLFunc: tt.mockShortenURLFunc,
			}

			handler := &Handler{
				urlService: mockService,
			}

			handler.HandleShortenJSON(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedResponse != nil {
				var response ShortenResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Result != tt.expectedResponse.Result {
					t.Errorf("Expected result %s, got %s", tt.expectedResponse.Result, response.Result)
				}

				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}

func TestShortenRequestUnmarshal(t *testing.T) {
	jsonStr := `{"url":"https://practicum.yandex.ru"}`
	var req ShortenRequest

	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if req.URL != "https://practicum.yandex.ru" {
		t.Errorf("Expected URL to be 'https://practicum.yandex.ru', got '%s'", req.URL)
	}
}

func TestShortenResponseMarshal(t *testing.T) {
	resp := ShortenResponse{
		Result: "http://localhost:8080/abc123",
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	expectedJSON := `{"result":"http://localhost:8080/abc123"}`
	if string(jsonBytes) != expectedJSON {
		t.Errorf("Expected JSON '%s', got '%s'", expectedJSON, string(jsonBytes))
	}
}
