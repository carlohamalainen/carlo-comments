package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/smithy-go"
	"github.com/carlohamalainen/carlo-comments/conduit"
)

const PostIndex = "PostIndex"

type CommentService struct {
	*DB
	DynamoDBRegion    string
	DynamoDBTableName string
}

type DynamoComment struct {
	SiteID        string `dynamodbav:"SiteID"`
	CommentID     string `dynamodbav:"CommentID"`
	PostID        string `dynamodbav:"PostID"`
	Timestamp     int64  `dynamodbav:"Timestamp"`
	SourceAddress string `dynamodbav:"SourceAddress"`
	Author        string `dynamodbav:"Author"`
	AuthorEmail   string `dynamodbav:"AuthorEmail"`
	IsActive      int    `dynamodbav:"IsActive"`
	CommentBody   string `dynamodbav:"CommentBody"`
}

// Two isomorphisms:
// 1. Timestamp: time.Time <=> UnixMilli
// 2. IsActive: 0, 1 <=> False, True
func commentToDynamoItem(c conduit.Comment) DynamoComment {
	isActive := 0
	if c.IsActive {
		isActive = 1
	}

	return DynamoComment{
		SiteID:        c.SiteID,
		CommentID:     c.CommentID,
		PostID:        c.PostID,
		Timestamp:     time.Time(c.Timestamp).UnixMilli(),
		SourceAddress: c.SourceAddress,
		Author:        c.Author,
		AuthorEmail:   c.AuthorEmail,
		IsActive:      isActive,
		CommentBody:   c.CommentBody,
	}
}

func dynamoItemToComment(d DynamoComment) conduit.Comment {
	return conduit.Comment{
		SiteID:        d.SiteID,
		CommentID:     d.CommentID,
		PostID:        d.PostID,
		Timestamp:     conduit.Timestamp(time.UnixMilli(d.Timestamp)),
		SourceAddress: d.SourceAddress,
		Author:        d.Author,
		AuthorEmail:   d.AuthorEmail,
		IsActive:      d.IsActive == 1,
		CommentBody:   d.CommentBody,
	}

}

func expandAWSError(err error, operation string) (string, []any) {
	msg := ""
	attrs := make([]any, 0)

	var ae smithy.APIError
	if errors.As(err, &ae) {
		attrs = append(attrs, slog.String("operation", operation))
		attrs = append(attrs, slog.String("error_code", ae.ErrorCode()))
		attrs = append(attrs, slog.String("error_message", ae.ErrorMessage()))
		attrs = append(attrs, slog.String("error_fault", ae.ErrorFault().String()))

		msg = "DynamoDB operation failed"
	} else {
		attrs = append(attrs, slog.String("operation", operation))
		attrs = append(attrs, slog.String("error_type", reflect.TypeOf(err).String()))
		attrs = append(attrs, slog.String("error", err.Error()))
		msg = "Unknown DynamoDB error"
	}

	return msg, attrs
}

func NewCommentService(db *DB, dynamodbRegion string, dynamoDBTableName string) *CommentService {
	return &CommentService{db, dynamodbRegion, dynamoDBTableName}
}

func (cs *CommentService) NrComments(ctx context.Context, filter conduit.CommentFilter) (int, error) {
	logger := conduit.GetLogger(ctx)

	if filter.SiteID == nil {
		return -1, fmt.Errorf("need SiteID for count query")
	}
	if filter.PostID == nil {
		return -1, fmt.Errorf("need PostID for count query")
	}

	query := &dynamodb.QueryInput{
		TableName:              aws.String(cs.DynamoDBTableName),
		IndexName:              aws.String(PostIndex),
		KeyConditionExpression: aws.String("SiteID = :siteID AND PostID = :postID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":siteID": &types.AttributeValueMemberS{Value: *filter.SiteID},
			":postID": &types.AttributeValueMemberS{Value: *filter.PostID},
		},
	}

	result, err := cs.Client.Query(ctx, query)
	if err != nil {
		msg, attrs := expandAWSError(err, "query comments")
		logger.ErrorContext(ctx, msg, attrs...)
		return -1, err
	}

	var comments []DynamoComment
	err = attributevalue.UnmarshalListOfMaps(result.Items, &comments)
	if err != nil {
		msg, attrs := expandAWSError(err, "unmarshall")
		logger.ErrorContext(ctx, msg, attrs...)
	}

	return len(comments), nil
}

func (cs *CommentService) UpsertComment(ctx context.Context, c *conduit.Comment) error {
	logger := conduit.GetLogger(ctx)

	item, err := attributevalue.MarshalMap(commentToDynamoItem(*c))
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %v", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(cs.DynamoDBTableName),
	}

	_, err = cs.Client.PutItem(ctx, input)
	if err != nil {
		msg, attrs := expandAWSError(err, "PutItem")
		logger.ErrorContext(ctx, msg, attrs...)
		return fmt.Errorf("failed to put item in DynamoDB: %v due to %v", err, attrs)
	}

	return nil
}

func (cs *CommentService) Comments(ctx context.Context, commentFilter conduit.CommentFilter) ([]conduit.Comment, error) {
	logger := conduit.GetLogger(ctx)
	var empty []conduit.Comment

	if commentFilter.SiteID == nil {
		return empty, fmt.Errorf("need SiteID for Comment query")
	}

	var query *dynamodb.QueryInput

	switch {

	// With a CommentID, we search on that and ignore any PostID in the filter.
	// This would be so much nicer as an ADT.
	case commentFilter.CommentID != nil:
		query = &dynamodb.QueryInput{
			TableName:              aws.String(cs.DynamoDBTableName),
			KeyConditionExpression: aws.String("SiteID = :siteID AND CommentID = :commentID"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":siteID":    &types.AttributeValueMemberS{Value: *commentFilter.SiteID},
				":commentID": &types.AttributeValueMemberS{Value: *commentFilter.CommentID},
			},
		}

	// With a PostID, we search for SiteID + PostID using a secondary global index.
	case commentFilter.PostID != nil:
		query = &dynamodb.QueryInput{
			TableName:              aws.String(cs.DynamoDBTableName),
			IndexName:              aws.String(PostIndex),
			KeyConditionExpression: aws.String("SiteID = :siteID AND PostID = :postID"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":siteID": &types.AttributeValueMemberS{Value: *commentFilter.SiteID},
				":postID": &types.AttributeValueMemberS{Value: *commentFilter.PostID},
			},
		}
	default:
		logger.Error("unhandled case in DynamoDB Comments", "filter", commentFilter) // FIXME log this better
		return empty, fmt.Errorf("unhandled case in DynamoDB Comments")
	}

	// Regardless of the query, we offer to filter on IsActive.
	if commentFilter.IsActive != nil {
		a := "0"
		if *commentFilter.IsActive {
			a = "1"
		}
		query.FilterExpression = aws.String("IsActive = :isActive")
		query.ExpressionAttributeValues[":isActive"] = &types.AttributeValueMemberN{Value: a}
	}

	result, err := cs.Client.Query(ctx, query)
	if err != nil {
		msg, attrs := expandAWSError(err, "query comments")
		logger.ErrorContext(ctx, msg, attrs...)
		return empty, err
	}

	var dynamoComments []DynamoComment
	err = attributevalue.UnmarshalListOfMaps(result.Items, &dynamoComments)
	if err != nil {
		msg, attrs := expandAWSError(err, "unmarshall")
		logger.ErrorContext(ctx, msg, attrs...)
		return empty, err
	}

	var comments []conduit.Comment
	for _, d := range dynamoComments {
		comments = append(comments, dynamoItemToComment(d))
	}

	return comments, nil
}

func (cs *CommentService) DeleteComment(ctx context.Context, comment *conduit.Comment) error {
	logger := conduit.GetLogger(ctx)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(cs.DynamoDBTableName),
		Key: map[string]types.AttributeValue{
			"SiteID":    &types.AttributeValueMemberS{Value: comment.SiteID},
			"CommentID": &types.AttributeValueMemberS{Value: comment.CommentID},
		},
	}
	_, err := cs.Client.DeleteItem(ctx, input)
	msg, attrs := expandAWSError(err, "DeleteItem")
	logger.ErrorContext(ctx, msg, attrs...)

	return err
}
