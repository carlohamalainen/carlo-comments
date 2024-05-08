package server

import (
	"github.com/google/uuid"
	"net/http"

	"github.com/carlohamalainen/carlo-comments/conduit"
)

func (s *Server) loginUser() http.HandlerFunc {
	type Input struct {
		User struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.Logger.With("request_id", uuid.NewString(), "handler", "loginUser")
		ctx := conduit.WithLogger(r.Context(), logger)

		input := Input{}

		if err := readJSON(ctx, r.Body, &input, s.Config.MaxBodySize); err != nil {
			logger.Error("failed to decode json", "error", err)
			errorResponse(ctx, w, http.StatusUnprocessableEntity, err)
			return
		}

		user, err := s.UserService.Authenticate(ctx, s.Config.AdminUser, s.Config.AdminPass, input.User.Email, input.User.Password)

		if err != nil || user == nil {
			invalidUserCredentialsError(ctx, w)
			return
		}

		type UserResponse struct {
			Email string `json:"email"`
			Token string `json:"token"`
		}

		writeJSON(ctx, w, http.StatusOK, M{"user": UserResponse{Email: user.Email, Token: user.Token}})
	}
}
