package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lehigh-university-libraries/isle-fcrepo-fs/handler"
)

func main() {
	if os.Getenv("OCFL_ROOT") == "" {
		slog.Error("OCFL_ROOT is required. e.g. OCFL_ROOT=/opt/ocfl")
		os.Exit(1)
	}

	// create a healthcheck with no middleware/auth
	r := mux.NewRouter()
	r.HandleFunc("/healthcheck", handler.HealthCheck).Methods("GET")

	// create the main route with logging and JWT auth middleware
	authRouter := r.PathPrefix("/").Subrouter()
	authRouter.Use(handler.LoggingMiddleware, handler.AuthMiddleware)
	authRouter.PathPrefix("/").HandlerFunc(handler.Download)

	port := "8080"
	slog.Info("Server is starting", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("Server failed", "err", err)
		os.Exit(1)
	}
}
