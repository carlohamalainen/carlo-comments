package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	"github.com/carlohamalainen/carlo-comments/conduit"
)

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

	keyCond := expression.Key("SiteID").Equal(expression.Value(*filter.SiteID))
	filt := expression.Name("PostID").Equal(expression.Value(*filter.PostID))

	builder := expression.NewBuilder().WithKeyCondition(keyCond)
	builder = builder.WithFilter(filt)
	expr, err := builder.Build()
	if err != nil {
		logger.Error("DynamoDB build failed", "error", err, "filter", filter)
		return -1, err
	}

	query := &dynamodb.QueryInput{
		TableName:                 aws.String(cs.DynamoDBTableName),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := cs.DynamoDB.Query(query)
	if err != nil {
		logger.Error("failed DynamoDB query", "error", err, "query", query)
		return -1, err
	}

	var comments []DynamoComment
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &comments)
	if err != nil {
		logger.Error("unmarshal DynamoDB result failed", "error", err)
		return -1, err
	}

	return len(comments), nil
}

func (cs *CommentService) UpsertComment(ctx context.Context, c *conduit.Comment) error {
	logger := conduit.GetLogger(ctx)

	item, err := dynamodbattribute.MarshalMap(commentToDynamoItem(*c))
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %v", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(cs.DynamoDBTableName),
	}

	_, err = cs.DynamoDB.PutItem(input)
	if err != nil {
		logger.Error("DynamoDB put failed", "error", err)
		return fmt.Errorf("failed to put item in DynamoDB: %v", err)
	}

	return nil
}

func (cs *CommentService) Comments(ctx context.Context, commentFilter conduit.CommentFilter) ([]conduit.Comment, error) {
	logger := conduit.GetLogger(ctx)

	var empty []conduit.Comment

	if commentFilter.SiteID == nil {
		return empty, fmt.Errorf("need SiteID for count query")
	}

	keyCond := expression.Key("SiteID").Equal(expression.Value(*commentFilter.SiteID))

	var conditions []expression.ConditionBuilder

	if commentFilter.PostID != nil {
		conditions = append(conditions, expression.Name("PostID").Equal(expression.Value(*commentFilter.PostID)))
	}

	if commentFilter.IsActive != nil {
		conditions = append(conditions, expression.Name("IsActive").Equal(expression.Value(*commentFilter.IsActive)))
	}

	var filterCond expression.ConditionBuilder
	if len(conditions) > 0 {
		filterCond = conditions[0]
		for _, cond := range conditions[1:] {
			filterCond = filterCond.And(cond)
		}
	}

	builder := expression.NewBuilder().WithKeyCondition(keyCond)
	if len(conditions) > 0 {
		builder = builder.WithFilter(filterCond)
	}

	expr, err := builder.Build()
	if err != nil {
		logger.Error("DynamoDB build failed", "error", err)
		return empty, err
	}

	query := &dynamodb.QueryInput{
		TableName:                 aws.String(cs.DynamoDBTableName),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	result, err := cs.DynamoDB.Query(query)
	if err != nil {
		logger.Error("DynamoDB query failed", "error", err, "query", query)
		return empty, err
	}

	var dynamoComments []DynamoComment
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &dynamoComments)
	if err != nil {
		logger.Error("unmarshal from DynamoDB failed", "error", err)
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
		Key: map[string]*dynamodb.AttributeValue{
			"SiteID": { S: aws.String(comment.SiteID), },
			"CommentID": { S: aws.String(comment.CommentID), },
		},
		ConditionExpression: aws.String("attribute_exists(CommentID)"),
	}

	_, err := cs.DynamoDB.DeleteItem(input)
	if err != nil {
		if aerr, ok := err.(*dynamodb.ConditionalCheckFailedException); ok {
			logger.Error("item does not exist or condition not met", "error", aerr.Message())
		} else {
			logger.Error("error deleting item", "error", err.Error())
		}
		return err
	}

	return nil
}
