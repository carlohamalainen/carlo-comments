package server

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	"github.com/carlohamalainen/carlo-comments/conduit"
	carloconfig "github.com/carlohamalainen/carlo-comments/config"
)

func dynamoUrl(cfg *carloconfig.Config, comment *conduit.Comment) string {
	// FIXME Each backend should make their own link.

	url := "https://"
	url += cfg.DynamoDBRegion + ".console.aws.amazon.com/dynamodbv2/home?region=" + cfg.DynamoDBRegion
	url += "#edit-item?itemMode=2&pk="
	url += comment.SiteID
	url += "&route=ROUTE_ITEM_EXPLORER&sk="
	url += comment.CommentID
	url += "&table="
	url += cfg.DynamoDBTableName

	return url
}

func inactiveSearch(cfg *carloconfig.Config) string {
	url := "https://"
	url += cfg.DynamoDBRegion + ".console.aws.amazon.com/dynamodbv2/home?region=" + cfg.DynamoDBRegion
	url += "#item-explorer?filter1Comparator=EQUAL&filter1Name=IsActive&filter1Type=N&filter1Value=0&operation=SCAN&table="
	url += cfg.DynamoDBTableName

	return url
}

type EmailData struct {
	DynamoDBLink    string
	AllInactiveLink string
	Comment         conduit.Comment
}

func Notify(logger *slog.Logger, cfg *carloconfig.Config, comment *conduit.Comment) error {
	ctx := context.Background()

	awscfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.DynamoDBRegion))
	if err != nil {
		return err
	}

	client := ses.NewFromConfig(awscfg)

	sender := cfg.AdminUser
	recipient := cfg.AdminUser
	subject := "New comment " + comment.CommentID + " " + comment.PostID

	data := EmailData{
		Comment:         *comment,
		DynamoDBLink:    dynamoUrl(cfg, comment),
		AllInactiveLink: inactiveSearch(cfg),
	}

	textBody := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
		comment.CommentID,
		comment.PostID,
		comment.Author,
		comment.AuthorEmail,
		comment.CommentBody,
		data.DynamoDBLink,
		data.AllInactiveLink)

	emailTemplate := `
<!DOCTYPE html>
<html>
<body>
	<p>{{.Comment.CommentID}}</p>
	<p>{{.Comment.PostID}}</p>
	<p>{{.Comment.Author}}</p>
	<p>{{.Comment.AuthorEmail}}</p>
	<p>{{.Comment.CommentBody}}</p>
	<p><a href="{{.DynamoDBLink}}">{{.DynamoDBLink}}</a></p>
	<p><a href="{{.AllInactiveLink}}">{{.AllInactiveLink}}</a></p>
</body>
</html>
`

	tmpl, err := template.New("emailTemplate").Parse(emailTemplate)
	if err != nil {
		logger.Error("failed to parse email template", "error", err.Error())
		return err
	}

	var bodyBuffer bytes.Buffer
	if err := tmpl.Execute(&bodyBuffer, data); err != nil {
		logger.Error("failed to execute email template", "error", err.Error())
		return err
	}

	htmlBody := bodyBuffer.String()

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(htmlBody),
				},
				Text: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(textBody),
				},
			},
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	result, err := client.SendEmail(ctx, input)
	if err != nil {
		logger.Error("failed to send email", "error", err.Error())
		return err
	}

	logger.Info("sent email notification", "recipient", recipient, "message_id", result.MessageId)

	return nil
}
