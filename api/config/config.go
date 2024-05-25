package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// sqlite config
	SqlitePath string

	// S3 config
	S3Region     string
	S3BucketName string

	Port               string
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

	MaxNrComments int

	LimiterRate  int
	LimiterBurst int
}

func GetConfig() (*Config, error) {
	cfg := &Config{MaxNrComments: 100}

	sqlitePath, ok := os.LookupEnv("SQLITE_PATH")
	if ok {
		cfg.SqlitePath = sqlitePath
	}

	s3BucketName, ok := os.LookupEnv("S3_BUCKET")
	if ok {
		cfg.S3BucketName = s3BucketName
	}
	s3Region, ok := os.LookupEnv("S3_REGION")
	if ok {
		cfg.S3Region = s3Region
	}

	storageOK := false
	if cfg.SqlitePath != "" && cfg.S3BucketName == "" && cfg.S3Region == "" {
		storageOK = true
	}
	if cfg.SqlitePath == "" && cfg.S3BucketName != "" && cfg.S3Region != "" {
		storageOK = true
	}
	if !storageOK {
		return nil, fmt.Errorf("need sqlite or S3 environment")
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		return nil, fmt.Errorf("PORT is not set")
	}
	cfg.Port = port

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

	limiterRate, ok := os.LookupEnv("LIMITER_RATE")
	if !ok {
		return nil, fmt.Errorf("LIMITER_RATE is not set")
	}
	num, err := strconv.ParseInt(limiterRate, 10, strconv.IntSize)
	if err != nil {
		return nil, fmt.Errorf("LIMITER_RATE bad integer")

	}
	cfg.LimiterRate = int(num)

	limiterBurst, ok := os.LookupEnv("LIMITER_BURST")
	if !ok {
		return nil, fmt.Errorf("LIMITER_BURST is not set")
	}
	num, err = strconv.ParseInt(limiterBurst, 10, strconv.IntSize)
	if err != nil {
		return nil, fmt.Errorf("LIMITER_BURST bad integer")

	}
	cfg.LimiterBurst = int(num)

	cfg.MaxBodySize = 8192

	return cfg, nil
}
