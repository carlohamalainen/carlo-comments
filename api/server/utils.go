package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/golang-jwt/jwt"
	"github.com/microcosm-cc/bluemonday"
)

type M map[string]interface{}

func writeJSON(ctx context.Context, w http.ResponseWriter, code int, data interface{}) {
	logger := conduit.GetLogger(ctx)

	jsonBytes, err := json.Marshal(data)

	if err != nil {
		logger.Error("marshal JSON failed", "error", err)
		serverError(ctx, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(jsonBytes)

	if err != nil {
		logger.Error("write JSON response failed", "error", err)
	}
}

func readJSON(ctx context.Context, body io.Reader, input interface{}, maxLength int) error {
	logger := conduit.GetLogger(ctx)

	if maxLength > 0 {
		buf := make([]byte, maxLength)

		n, err := io.ReadFull(body, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("body exceeds maximum length %v", maxLength)
		}
		buf = buf[:n]
		
		err = json.Unmarshal(buf, input)
		if err != nil {
			logger.Debug("failed to decode JSON", "error", err, "body", string(buf))
			return fmt.Errorf("failed to decode JSON: %v", err)
		}
		return nil
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}

	err = json.Unmarshal(data, input)
	if err != nil {
		logger.Debug("failed to decode JSON", "error", err, "body", string(data))
	}
	return err
}

func Sanitize(comment string) string {
	p := bluemonday.UGCPolicy()
	sanitized := p.Sanitize(comment)
	return sanitized
}

func (s *Server) GenerateUserToken(ctx context.Context, user *conduit.User) (string, error) {
	logger := conduit.GetLogger(ctx)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   user.Email,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(s.Config.HmacSecret)

	if err != nil {
		logger.Error("failed to generate JWT token", "error", err)
		return "", err
	}

	return tokenString, nil
}
