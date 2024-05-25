package main

import (
	// "fmt"
	"context"
	"fmt"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/carlohamalainen/carlo-comments/config"
	"github.com/carlohamalainen/carlo-comments/s3"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	close, logger := conduit.NewLogger(*cfg)
	defer close()

	ctx := conduit.WithLogger(context.Background(), logger)

	db, err := s3.Open(ctx, *cfg)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	cs := s3.NewCommentService(db, cfg.S3Region, cfg.S3BucketName)

	fmt.Println(cs)

	siteID := "carlo-hamalainen.net"
	postID := "/2024/01/01/foo"
	nr, err := cs.NrComments(ctx, conduit.CommentFilter{SiteID: &siteID, PostID: &postID})
	if err != nil {
		panic(err)
	}

	fmt.Println(nr)

	// us := simple.NewUserService(cfg.HmacSecret)

	// fmt.Printf("Password: ")
	// bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(string(us.HashPassword(bytePassword)))
}
