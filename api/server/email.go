package server

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/carlohamalainen/carlo-comments/config"
)

func Notify(config *config.Config, comment *conduit.Comment) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.S3Region)},
	)
	if err != nil {
		return err
	}

	svc := ses.New(sess)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(config.AdminUser),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Data: aws.String(comment.CommentID + "\n" + comment.PostID + "\n" + comment.Author + "\n" + comment.AuthorEmail + "\n" + comment.CommentBody),
				},
			},
			Subject: &ses.Content{
				Data: aws.String("New comment " + comment.CommentID + " " + comment.PostID),
			},
		},
		Source:    aws.String(config.AdminUser),
		SourceArn: aws.String(config.SESIdentity),
	}

	// result, err := svc.SendEmail(input)
	_, err = svc.SendEmail(input)
	if err != nil {
		return err
	}

	// fmt.Println("Email sent successfully:", result.MessageId)

	return nil
}
