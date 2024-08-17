package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/carlohamalainen/carlo-comments/config"
)

type DB struct {
	*dynamodb.DynamoDB
}

func Open(ctx context.Context, cfg config.Config) (*DB, error) {
	logger := conduit.GetLogger(ctx)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.S3Region),
	})
	if err != nil {
		logger.Error("failed to create AWS session", "error", err)
		return nil, err
	}

	if _, err := sess.Config.Credentials.Get(); err != nil {
		logger.Error("authentication failed", "error", err)
	}

	return &DB{dynamodb.New(sess)}, nil
}
