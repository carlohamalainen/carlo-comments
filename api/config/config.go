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

	SESIdentity  string

	// S3 config
	S3Region     string
	S3BucketName string

	// DynamoDB config
	DynamoDBRegion     string
	DynamoDBTableName string

	// CloudFlare
	CfSiteKey   string
	CfSecretKey string

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

func setDynamoDBConfig(config *Config) int {
	dynamoDBTableName, ok0 := os.LookupEnv("DYNAMODB_TABLE_NAME")
	dynamoDBRegion, ok1 := os.LookupEnv("DYNAMODB_REGION")

	if ok0 && ok1 {
		config.DynamoDBTableName = dynamoDBTableName
		config.DynamoDBRegion = dynamoDBRegion
		return 1
	}

	return 0
}

func setSQLiteConfig(config *Config) int {
	sqlitePath, ok := os.LookupEnv("SQLITE_PATH")
	if ok {
		config.SqlitePath = sqlitePath
		return 1
	}
	return 0
}

func setS3Config(config *Config) int {
	s3BucketName, ok0 := os.LookupEnv("S3_BUCKET")
	s3Region, ok1 := os.LookupEnv("S3_REGION")
	if ok0 && ok1 {
		config.S3BucketName = s3BucketName
		config.S3Region = s3Region
		return 1
	}

	return 0
}


func GetConfig() (*Config, error) {
	cfg := &Config{MaxNrComments: 100}

	dynamodb := setDynamoDBConfig(cfg)
	s3 := setS3Config(cfg)
	sqlite := setSQLiteConfig(cfg)

	if dynamodb + s3 + sqlite != 1 {
		return nil, fmt.Errorf("need precisely one backend to be configured")

	}

	sesIdentity, ok := os.LookupEnv("SES_IDENTITY")
	if !ok {
		return nil, fmt.Errorf("SES_IDENTITY is not set")
	}
	cfg.SESIdentity = sesIdentity

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

	cfg.MaxBodySize = 4 * 8192

	cfSiteKey, ok := os.LookupEnv("CF_SITE_KEY")
	if !ok {
		return nil, fmt.Errorf("CF_SITE_KEY is not set")
	}
	cfg.CfSiteKey = cfSiteKey

	cfSecretKey, ok := os.LookupEnv("CF_SECRET_KEY")
	if !ok {
		return nil, fmt.Errorf("CF_SECRET_KEY is not set")
	}
	cfg.CfSecretKey = cfSecretKey

	return cfg, nil
}
