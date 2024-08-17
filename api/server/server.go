package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/carlohamalainen/carlo-comments/config"
	"github.com/carlohamalainen/carlo-comments/dynamodb"
	"github.com/carlohamalainen/carlo-comments/simple"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

// TODO import/export for admin

func (s *Server) limitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type Server struct {
	server *http.Server

	router  *mux.Router
	limiter *rate.Limiter

	Config config.Config

	UserService    conduit.UserService
	commentService conduit.CommentService

	logLevel slog.Level

	Logger *slog.Logger

	mtx        sync.Mutex
	knownPosts map[string](map[string]bool)
}

func (s *Server) InitState() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.knownPosts = make(map[string]map[string]bool)
}

func (s *Server) Count() int {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	nr := 0

	for siteID := range s.knownPosts {
		for range s.knownPosts[siteID] {
			nr = nr + 1
		}
	}

	return nr
}

func (s *Server) IsKnown(site_id string, post_id string) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	siteMap, ok := s.knownPosts[site_id]
	if !ok {
		return false
	}

	_, ok = siteMap[post_id]
	return ok
}

func (s *Server) SetKnown(ctx context.Context, site_id string, post_id string) {
	logger := conduit.GetLogger(ctx)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	_, ok := s.knownPosts[site_id]
	if !ok {
		s.knownPosts[site_id] = make(map[string]bool)
	}

	logger.Info("setting known host", "site_id", site_id, "post_id", post_id)
	s.knownPosts[site_id][post_id] = true
}

func NewServer(ctx context.Context, db *dynamodb.DB, cfg config.Config) *Server {
	logger := conduit.GetLogger(ctx)

	s := Server{
		server: &http.Server{
			WriteTimeout: cfg.HandlerTimeout,
			ReadTimeout:  cfg.HandlerTimeout,
			IdleTimeout:  cfg.HandlerTimeout,
		},

		router: mux.NewRouter().StrictSlash(true),

		limiter:  rate.NewLimiter(rate.Limit(cfg.LimiterRate), cfg.LimiterBurst),
		logLevel: cfg.LogLevel,

		Logger: logger,

		Config: cfg,
	}

	s.routes()

	err := s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			methods, err := route.GetMethods()
			if err == nil {
				for _, method := range methods {
					logger.Info("registered path", "path_template", pathTemplate, "method", method)
				}
			}
		}
		return nil
	})

	if err != nil {
		logger.Error("error walking the router", "error", err)
	}

	s.UserService = simple.NewUserService(s.Config.HmacSecret)
	s.commentService = dynamodb.NewCommentService(db, cfg.DynamoDBRegion, cfg.DynamoDBTableName)

	// Maybe State should be a conduit as well, with an in-memory thing...
	s.InitState()

	s.server.Handler = s.router

	return &s
}

func (s *Server) Run(ctx context.Context, port string) error {
	logger := conduit.GetLogger(ctx)

	s.server.Addr = port

	logger.Info("starting server", "addr", s.server.Addr)

	return s.server.ListenAndServe()
}

func (s *Server) healthCheck() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.Logger.With("request_id", uuid.NewString(), "handler", "healthCheck")
		ctx := conduit.WithLogger(r.Context(), logger)

		nr := s.Count()

		resp := make(map[string]interface{})
		data := make(map[string]interface{})

		resp["status"] = "available"
		resp["message"] = "healthy"
		resp["now"] = time.Now()

		data["count"] = nr

		resp["data"] = data

		logger.Info("health check", "handler", "healthCheck", "count", nr)

		writeJSON(ctx, w, http.StatusOK, resp)
	})
}
