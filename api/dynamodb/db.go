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

	sess, err := session.NewSession()
	if err != nil {
		logger.Error("failed to create AWS session", "error", err)
		return nil, err
	}

	if _, err := sess.Config.Credentials.Get(); err != nil {
		logger.Error("authentication failed", "error", err)
	}

	svc := dynamodb.New(sess, aws.NewConfig().WithRegion(cfg.DynamoDBRegion))

	return &DB{svc}, nil
}
