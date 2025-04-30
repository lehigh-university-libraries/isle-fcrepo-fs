package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lehigh-university-libraries/isle-fcrepo-fs/handler"
)

func main() {
	if os.Getenv("JWKS_URI") == "" {
		slog.Error("JWKS_URI is required. e.g. JWKS_URI=https://gitlab.com/oauth/discovery/keys")
		os.Exit(1)
	}
	err := handler.FetchJWKS()
	if err != nil {
		slog.Error("Unable to fetch JWKS", "uri", os.Getenv("JWKS_URI"), "err", err)
		os.Exit(1)
	}

	if os.Getenv("OCFL_ROOT") == "" {
		slog.Error("OCFL_ROOT is required. e.g. OCFL_ROOT=/opt/ocfl")
		os.Exit(1)
	}

	// create a healthcheck with no middleware/auth
	r := mux.NewRouter()
	r.HandleFunc("/healthcheck", handler.HealthCheck).Methods("GET")

	// create the main route with logging and JWT auth middleware
	authRouter := r.PathPrefix("/").Subrouter()
	authRouter.Use(handler.LoggingMiddleware, handler.JWTAuthMiddleware)
	authRouter.PathPrefix("/").HandlerFunc(handler.Download)

	port := "8080"
	slog.Info("Server is starting", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("Server failed", "err", err)
		os.Exit(1)
	}
}
