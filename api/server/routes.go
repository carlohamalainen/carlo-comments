package server

import (
	"fmt"
	"log/slog"

	"github.com/rs/cors"
)

type FilterMode int

const (
	ActiveOnly FilterMode = iota // query must be SiteID, PostID; only returns active
	FreeRange                    // query can be anything at all (only for admin interface)
)

type printfLogger struct{
	slog *slog.Logger
}

func (l *printfLogger) Printf(format string, args ...interface{}) {
	l.slog.Info("cors", "message", fmt.Sprintf(format, args...))
}

func (s *Server) routes() {
	cors := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins: s.Config.CorsAllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// AllowedHeaders: []string{"Content-Type", "Authorization"},
		AllowedHeaders: []string{"*"},
		Logger: &printfLogger{slog:s.Logger},
	})
	s.router.Use(cors.Handler)
	s.router.Use(s.limitMiddleware)
	s.router.Use(Logger(s.Logger))

	v1 := s.router.PathPrefix("/v1").Subrouter()

	noAuth := v1.PathPrefix("").Subrouter()
	{
		noAuth.Handle("/health", s.healthCheck()).Methods("GET") // FIXME healthz as per convention; kubernetes config change too

		// Need OPTIONS here otherwise the cors handler won't match anything!
		noAuth.Handle("/comments/new", s.createComment()).Methods("POST", "OPTIONS")
		noAuth.Handle("/comments", s.getComments(ActiveOnly)).Methods("POST", "OPTIONS")
	}

	admin := v1.PathPrefix("/admin").Subrouter()
	{
		admin.Handle("/login", s.loginUser()).Methods("POST")
	}

	comments := admin.PathPrefix("/comments").Subrouter()
	comments.Use(s.authenticate())
	{
		comments.Handle("/new", s.upsertComment()).Methods("POST")
		comments.Handle("", s.getComments(FreeRange)).Methods("POST")
	}
}
