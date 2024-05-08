package server

import (
	"context"
	"net/http"
)

func badRequestError(ctx context.Context, w http.ResponseWriter) {
	errorResponse(ctx, w, http.StatusUnprocessableEntity, "unable to process request")
}

func invalidUserCredentialsError(ctx context.Context, w http.ResponseWriter) {
	msg := "invalid authentication credentials"
	errorResponse(ctx, w, http.StatusUnauthorized, msg)
}

func serverError(ctx context.Context, w http.ResponseWriter) {
	errorResponse(ctx, w, http.StatusInternalServerError, "internal error")
}

func errorResponse(ctx context.Context, w http.ResponseWriter, code int, errs interface{}) {
	writeJSON(ctx, w, code, M{"errors": errs})
}
