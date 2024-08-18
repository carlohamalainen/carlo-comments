package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/carlohamalainen/carlo-comments/conduit"
	carloconfig "github.com/carlohamalainen/carlo-comments/config"
)

type DB struct {
	*dynamodb.Client
}

func Open(ctx context.Context, carloconfig carloconfig.Config) (*DB, error) {
	logger := conduit.GetLogger(ctx)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(carloconfig.DynamoDBRegion))
	if err != nil {
		logger.Error("unable to load SDK config", "error", err.Error())
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DB{client}, nil
}
