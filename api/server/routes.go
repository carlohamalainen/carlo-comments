package server

import (
	"github.com/rs/cors"
)

type FilterMode int

const (
	ActiveOnly FilterMode = iota // query must be SiteID, PostID; only returns active
	FreeRange                    // query can be anything at all (only for admin interface)
)

func (s *Server) routes() {
	cors := cors.New(cors.Options{AllowedOrigins: s.Config.CorsAllowedOrigins})
	s.router.Use(s.limitMiddleware)
	s.router.Use(cors.Handler)
	s.router.Use(Logger(s.Logger))

	v1 := s.router.PathPrefix("/v1").Subrouter()

	noAuth := v1.PathPrefix("").Subrouter()
	{
		noAuth.Handle("/health", s.healthCheck()).Methods("GET")
		noAuth.Handle("/comments", s.createComment()).Methods("POST")
		noAuth.Handle("/comments", s.getComments(ActiveOnly)).Methods("GET")
	}

	admin := v1.PathPrefix("/admin").Subrouter()
	{
		admin.Handle("/login", s.loginUser()).Methods("POST")
	}

	comments := admin.PathPrefix("/comments").Subrouter()
	comments.Use(s.authenticate())
	{
		comments.Handle("", s.upsertComment()).Methods("POST")
		comments.Handle("", s.getComments(FreeRange)).Methods("GET")
	}
}
