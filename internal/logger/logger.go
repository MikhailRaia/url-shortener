package logger

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes default zerolog logger for the application.
func InitLogger() {
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)
}

// RequestLogger logs basic request/response metadata for each HTTP call.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := NewResponseWriter(w)

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		log.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Dur("duration", duration).
			Msg("Request processed")

		log.Info().
			Int("status", ww.Status()).
			Int("size", ww.Size()).
			Msg("Response sent")
	})
}

// ResponseWriter wraps http.ResponseWriter to capture status code and size.
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// NewResponseWriter creates a ResponseWriter wrapper.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Status returns the captured HTTP status code.
func (rw *ResponseWriter) Status() int {
	return rw.statusCode
}

// Size returns the total number of bytes written to the response.
func (rw *ResponseWriter) Size() int {
	return rw.size
}
