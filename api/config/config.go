package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port               string
	SqlitePath         string
	HmacSecret         string
	CommentHost        string
	AdminUser          string
	AdminPass          string
	LogLevel           slog.Level
	LogDirectory       string
	AppName            string
	CorsAllowedOrigins []string
	HandlerTimeout     time.Duration
	MaxBodySize        int
}

func GetConfig() (*Config, error) {
	cfg := &Config{}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		return nil, fmt.Errorf("PORT is not set")
	}
	cfg.Port = port

	sqlitePath, ok := os.LookupEnv("SQLITE_PATH")
	if !ok {
		return nil, fmt.Errorf("SQLITE_PATH is not set")
	}
	cfg.SqlitePath = sqlitePath

	hmacSecret, ok := os.LookupEnv("HMAC_SECRET")
	if !ok {
		return nil, fmt.Errorf("HMAC_SECRET is not set")
	}
	cfg.HmacSecret = hmacSecret

	commentHost, ok := os.LookupEnv("COMMENT_HOST")
	if !ok {
		return nil, fmt.Errorf("COMMENT_HOST is not set")
	}
	cfg.CommentHost = commentHost

	adminUser, ok := os.LookupEnv("ADMIN_USER")
	if !ok {
		return nil, fmt.Errorf("ADMIN_USER is not set")
	}
	cfg.AdminUser = adminUser

	adminPass, ok := os.LookupEnv("ADMIN_PASS")
	if !ok {
		return nil, fmt.Errorf("ADMIN_PASS is not set")
	}
	cfg.AdminPass = adminPass

	cfg.LogLevel = slog.LevelDebug // TODO make this an option

	logDirectory, ok := os.LookupEnv("LOG_DIRECTORY")
	if !ok {
		return nil, fmt.Errorf("LOG_DIRECTORY is not set")
	}
	cfg.LogDirectory = logDirectory

	appName, ok := os.LookupEnv("APP_NAME")
	if !ok {
		return nil, fmt.Errorf("APP_NAME is not set")
	}
	cfg.AppName = appName

	allowedOrigins, ok := os.LookupEnv("CORS_ALLOWED_ORIGINS")
	if !ok {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS is not set")
	}
	cfg.CorsAllowedOrigins = strings.Split(allowedOrigins, ",")

	handlerTimeoutString, ok := os.LookupEnv("HANDLER_TIMEOUT")
	if !ok {
		return nil, fmt.Errorf("HANDLER_TIMEOUT is not set")
	}

	handlerTimeoutDuration, err := time.ParseDuration(handlerTimeoutString)
	if err != nil {
		return nil, fmt.Errorf("error parsing duration: %v", err)
	}
	cfg.HandlerTimeout = handlerTimeoutDuration

	cfg.MaxBodySize = 8192

	return cfg, nil
}
