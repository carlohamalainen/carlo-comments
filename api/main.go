package main

import (
	"context"
	"time"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/carlohamalainen/carlo-comments/config"
	"github.com/carlohamalainen/carlo-comments/dynamodb"
	"github.com/carlohamalainen/carlo-comments/server"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	close, logger := conduit.NewLogger(*cfg)
	defer close()

	ctx := conduit.WithLogger(context.Background(), logger)

	db, err := dynamodb.Open(ctx, *cfg)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		panic(err)
	}

	srv := server.NewServer(ctx, db, *cfg)

	updater := func() {
		logger.Info("updating known hosts", "hosts", srv.Config.CommentHost)
		err := srv.InitHost(ctx, srv.Config.CommentHost)
		if err != nil {
			logger.Error("updater failed", "error", err)
		}
	}

	updater()

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			updater()
		}
	}()

	err = srv.Run(ctx, cfg.Port)
	if err != nil {
		logger.Error("server exited", "error", err)
	}
}
