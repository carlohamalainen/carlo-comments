package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/carlohamalainen/carlo-comments/conduit"
)

type CommentService struct {
	*DB
	S3Region     string
	S3BucketName string
}

func NewCommentService(db *DB, s3Region string, s3BucketName string) *CommentService {
	return &CommentService{db, s3Region, s3BucketName}
}

func (cs *CommentService) NrComments(ctx context.Context, filter conduit.CommentFilter) (int, error) {
	logger := conduit.GetLogger(ctx)

	var prefix = ""

	if filter.SiteID == nil {
		return 0, fmt.Errorf("need SiteID for count query")
	}
	if filter.PostID == nil {
		return 0, fmt.Errorf("need PostID for count query")
	}

	prefix += *filter.SiteID
	prefix += *filter.PostID
	prefix += "/"

	fmt.Println(*filter.SiteID)
	fmt.Println(prefix)
	resp, err := cs.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(cs.S3BucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		logger.Error("failed S3", "error", err, "action", "ListObjects", "prefix", prefix)
		return -1, err
	}

	return len(resp.Contents) - 1, nil
}

func (cs *CommentService) UpsertComment(ctx context.Context, c *conduit.Comment) error {
	logger := conduit.GetLogger(ctx)

	objectKey := c.SiteID + c.PostID + "/" + c.CommentID // FIXME sanity check the path?

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		logger.Error("json marshalling failure", "error", err)
		return err
	}

	_, err = cs.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(cs.S3BucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(jsonBytes),
	})
	if err != nil {
		logger.Error("failed S3", "error", err, "action", "PutObject")
		return err
	}

	return nil
}

func (cs *CommentService) Comments(ctx context.Context, commentFilter conduit.CommentFilter) ([]conduit.Comment, error) {
	logger := conduit.GetLogger(ctx)

	var empty []conduit.Comment
	var comments []conduit.Comment

	var prefix = ""

	// TODO this isn't consistent, need to have SiteID or SiteID+PostID

	if commentFilter.SiteID != nil && *commentFilter.SiteID != "" {
		prefix += *commentFilter.SiteID
	}

	if commentFilter.PostID != nil && *commentFilter.PostID != "" {
		prefix += *commentFilter.PostID
	}

	onlyActive := commentFilter.IsActive != nil && *commentFilter.IsActive

	resp, err := cs.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(cs.S3BucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		logger.Error("failed S3", "error", err, "action", "ListObjects", "prefix", prefix)
		return empty, err
	}

	for _, object := range resp.Contents {
		if *object.Key == prefix {
			continue
		}
		getResp, err := cs.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(cs.S3BucketName),
			Key:    object.Key,
		})
		if err != nil {
			logger.Error("failed S3", "error", err, "action", "GetObject", "key", *object.Key)
			return empty, err
		}

		var comment conduit.Comment
		err = json.NewDecoder(getResp.Body).Decode(&comment)
		if err != nil {
			logger.Error("failed json decode", "error", err, "key", *object.Key)
			return empty, err
		}

		if !onlyActive || (onlyActive && comment.IsActive) {
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

func (cs *CommentService) DeleteComment(ctx context.Context, comment *conduit.Comment) error {
	logger := conduit.GetLogger(ctx)

	key := comment.SiteID + comment.PostID + "/" + comment.CommentID

	_, err := cs.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(cs.S3BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Error("failed to DELETE comment", "error", err, "comment_id", comment.CommentID, "key", key)
		return err
	}

	return nil
}
