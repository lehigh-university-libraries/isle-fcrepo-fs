package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/httprc/v3/tracesink"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

var keySet jwk.Set

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

// JWTAuthMiddleware validates a JWT token and adds claims to the context
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Authorization")
		if a == "" || !strings.HasPrefix(strings.ToLower(a), "bearer ") {
			http.Error(w, "Missing Authorization header", http.StatusBadRequest)
			return
		}

		tokenString := a[7:]
		err := verifyJWT(tokenString)
		if err != nil {
			slog.Error("JWT verification failed", "err", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func verifyJWT(tokenString string) error {
	if keySet == nil {
		return fmt.Errorf("keySet not initialized")
	}
	// islandora will only ever provide a single key to sign JWTs
	// so just use the one key in JWKS
	key, ok := keySet.Key(0)
	if !ok {
		return fmt.Errorf("no key in jwks")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	if keySet.Len() > 1 {
		_, err = jwt.Parse([]byte(tokenString),
			jwt.WithContext(ctx),
			jwt.WithKeySet(keySet),
			jwt.WithVerify(true),
		)
	} else {
		_, err = jwt.Parse([]byte(tokenString),
			jwt.WithContext(ctx),
			jwt.WithKey(jwa.RS256(), key),
			jwt.WithVerify(true),
		)
	}

	if err != nil {
		return fmt.Errorf("unable to parse/verify token: %v", err)
	}

	return nil
}

// fetchJWKS fetches the JSON Web Key Set (JWKS) from the given URI
func FetchJWKS() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := jwk.NewCache(
		ctx,
		httprc.NewClient(
			httprc.WithTraceSink(tracesink.NewSlog(slog.New(slog.NewTextHandler(os.Stderr, nil)))),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create cache: %s", err)
	}

	jwksURI := os.Getenv("JWKS_URI")
	if err := c.Register(
		ctx,
		jwksURI,
		jwk.WithMaxInterval(24*time.Hour*7),
		jwk.WithMinInterval(24*time.Hour),
	); err != nil {
		return err
	}

	cached, err := c.CachedSet(jwksURI)
	if err != nil {
		return fmt.Errorf("failed to get cached keyset: %s", err)
	}
	keySet = cached

	return nil
}
