package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w GzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w GzipWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w GzipWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		contentType := wrapper.Header().Get("Content-Type")

		if strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html") ||
			strings.Contains(contentType, "text/plain") {

			if !wrapper.headersSent {
				gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
				if err != nil {
					w.WriteHeader(wrapper.statusCode)
					w.Write(wrapper.body)
					return
				}
				defer gz.Close()

				w.Header().Set("Content-Encoding", "gzip")

				for k, v := range wrapper.Header() {
					for _, vv := range v {
						w.Header().Add(k, vv)
					}
				}

				w.WriteHeader(wrapper.statusCode)

				gz.Write(wrapper.body)
			}
		} else {
			for k, v := range wrapper.Header() {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.WriteHeader(wrapper.statusCode)
			w.Write(wrapper.body)
		}
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode  int
	body        []byte
	headersSent bool
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func GzipReader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(w, r)
			return
		}

		gzReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "Failed to read gzipped request", http.StatusBadRequest)
			return
		}
		defer gzReader.Close()

		bodyReader := io.NopCloser(gzReader)
		r.Body = bodyReader
		r.ContentLength = -1

		next.ServeHTTP(w, r)
	})
}
