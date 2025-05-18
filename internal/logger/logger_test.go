package logger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitLogger(t *testing.T) {
	originalLogger := log.Logger
	originalStdout := os.Stdout

	defer func() {
		log.Logger = originalLogger
		os.Stdout = originalStdout
	}()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	InitLogger()

	assert.Equal(t, zerolog.InfoLevel, log.Logger.GetLevel())

	log.Info().Msg("test message")

	w.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)

	logStr := buf.String()
	assert.Contains(t, logStr, "time")
	assert.Contains(t, logStr, "message")
	assert.Contains(t, logStr, "test message")
}

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer

	originalLogger := log.Logger

	defer func() {
		log.Logger = originalLogger
	}()

	log.Logger = zerolog.New(&buf).Level(zerolog.InfoLevel)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	handler := RequestLogger(testHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response", rr.Body.String())

	logs := bytes.Split(buf.Bytes(), []byte("\n"))

	require.Equal(t, 3, len(logs), "Should have 2 log entries (plus an empty line)")

	var requestLog map[string]interface{}
	err := json.Unmarshal(logs[0], &requestLog)
	require.NoError(t, err)

	assert.Equal(t, "Request processed", requestLog["message"])
	assert.Equal(t, "GET", requestLog["method"])
	assert.Equal(t, "/test", requestLog["uri"])
	assert.Contains(t, requestLog, "duration")

	var responseLog map[string]interface{}
	err = json.Unmarshal(logs[1], &responseLog)
	require.NoError(t, err)

	assert.Equal(t, "Response sent", responseLog["message"])
	assert.Equal(t, float64(200), responseLog["status"])
	assert.Equal(t, float64(13), responseLog["size"])
}

func TestResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()

	rw := NewResponseWriter(rr)

	assert.Equal(t, http.StatusOK, rw.Status())
	assert.Equal(t, 0, rw.Size())

	rw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, rw.Status())

	n, err := rw.Write([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, rw.Size())

	n, err = rw.Write([]byte(" data"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 9, rw.Size())

	assert.Equal(t, "test data", rr.Body.String())
}
