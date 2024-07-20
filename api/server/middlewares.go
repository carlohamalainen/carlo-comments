package server

import (
	"net/http"
	"strings"

	// "github.com/carlohamalainen/carlo-comments/conduit"
	"log/slog"

	"github.com/golang-jwt/jwt"
)

func Logger(logger *slog.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.With(
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
				slog.String("proto", r.Proto),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("client_ip", getClientIP(r)),
			)

			logger.Info("Incoming request")

			// Create a new response writer to capture the response status code
			rw := &responseWriter{ResponseWriter: w}

			// Call the next handler
			logger.Info("calling the handler...")

			h.ServeHTTP(rw, r)

			// Log the response
			logger.With(slog.Int("status", rw.statusCode)).Info("Request completed")
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (s *Server) authenticate() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Authorization")
			authHeader := r.Header.Get("Authorization")

			authHeader = strings.Replace(authHeader, "Bearer ", "", 1)
			claims := &jwt.StandardClaims{}

			token, err := jwt.ParseWithClaims(authHeader, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(s.Config.HmacSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			r = setContextUserToken(r, token.Raw)
			h.ServeHTTP(w, r)
		})
	}
}
