package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// LoggingMiddleware logs incoming HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statusWriter := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(statusWriter, r)
		duration := time.Since(start)
		slog.Info(r.Method,
			"path", r.URL.Path,
			"status", statusWriter.statusCode,
			"duration", duration,
			"client_ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("ok"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Unable to write for healthcheck", "err", err)
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("https://%s/_flysystem/fedora%s", os.Getenv("DOMAIN"), r.URL.Path)
		req, err := http.NewRequest(http.MethodHead, url, nil)
		if err != nil {
			slog.Error("Unable to create request", "url", url, "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		for key, values := range r.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Not authorized", resp.StatusCode)
			return
		}

		next.ServeHTTP(w, r)
	})
}
