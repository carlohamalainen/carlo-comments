package conduit

import (
	"context"
	"net/mail"
	"strconv"
	"time"
)

type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {
	milliseconds := time.Time(t).UnixMilli()
	return []byte(strconv.FormatInt(milliseconds, 10)), nil
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	milliseconds, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*t = Timestamp(time.UnixMilli(milliseconds))
	return nil
}

type Comment struct {
	CommentID     string    `json:"commentID"`
	SiteID        string    `json:"siteID"`
	PostID        string    `json:"postID"`
	Timestamp     Timestamp `json:"timestamp"`
	SourceAddress string    `json:"sourceAddress"`
	Author        string    `json:"author"`
	AuthorEmail   string    `json:"authorEmail"`
	CommentBody   string    `json:"commentBody"`
	IsActive      bool      `json:"isActive"`
}

type NewComment struct {
	SiteID         string `json:"siteID"`
	PostID         string `json:"postID"`
	Author         string `json:"author"`
	AuthorEmail    string `json:"authorEmail"`
	CommentBody    string `json:"commentBody"`
	TurnstileToken string `json:"turnstileToken"`
}

type CommentFilter struct {
	CommentID *string
	SiteID    *string
	PostID    *string
	IsActive  *bool
}

type CommentService interface {
	NrComments(context.Context, CommentFilter) (int, error)
	UpsertComment(context.Context, *Comment) error
	Comments(context.Context, CommentFilter) ([]Comment, error)
	DeleteComment(context.Context, *Comment) error
}

// Validation utilities
//
// In a more general setting, try https://github.com/go-playground/validator/tree/master/_examples

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidPostID(key string) bool {
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '/' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
