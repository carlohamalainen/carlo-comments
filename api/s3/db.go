package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/carlohamalainen/carlo-comments/config"
)

type DB struct {
	*s3.S3
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

	awsCfg := sess.Config
	creds, err := awsCfg.Credentials.Get()
	if err != nil {
		// Authentication failed or credentials not found
		fmt.Println("Authentication failed:", err)
	}
	fmt.Println(creds)

	svc := s3.New(sess)

	return &DB{svc}, nil
}
