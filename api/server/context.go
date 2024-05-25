package server

// copied from github.com/0xdod/go-realworld

import (
	"context"
	"net/http"
)

type contextKey string

const (
	tokenKey contextKey = "carlo-comments-token"
)

func setContextUserToken(r *http.Request, token string) *http.Request {
	ctx := context.WithValue(r.Context(), tokenKey, token)
	return r.WithContext(ctx)
}
