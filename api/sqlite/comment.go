package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/carlohamalainen/carlo-comments/conduit"
)

type CommentService struct {
	*DB
}

func NewCommentService(db *DB) *CommentService {
	return &CommentService{db}
}

func (cs *CommentService) NrComments(ctx context.Context, filter conduit.CommentFilter) (int, error) {
	query := "SELECT COUNT(*) FROM comments WHERE site_id = ? AND post_id = ?"

	if filter.SiteID == nil {
		return 0, fmt.Errorf("need SiteID for count query")
	}
	if filter.PostID == nil {
		return 0, fmt.Errorf("need PostID for count query")
	}

	var count int
	err := cs.DB.QueryRow(query, filter.SiteID, filter.PostID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (cs *CommentService) UpsertComment(ctx context.Context, c *conduit.Comment) error {
	logger := conduit.GetLogger(ctx)

	upsert, err := cs.DB.Prepare(`
		INSERT OR REPLACE INTO comments (comment_id, site_id, post_id, timestamp, author, author_email, comment, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		logger.Error("prepare failed", "error", err)
		return err
	}
	defer upsert.Close()

	_, err = upsert.Exec(c.CommentID, c.SiteID, c.PostID, time.Time(c.Timestamp), c.Author, c.AuthorEmail, c.CommentBody, c.IsActive)
	if err != nil {
		logger.Error("exec failed", "error", err)
		return err
	}
	return nil
}

func (cs *CommentService) Comments(ctx context.Context, commentFilter conduit.CommentFilter) ([]conduit.Comment, error) {
	logger := conduit.GetLogger(ctx)

	var rows *sql.Rows
	var err error

	query := "SELECT comment_id, site_id, post_id, timestamp, author, author_email, comment, is_active FROM comments WHERE 1=1"
	args := []interface{}{}

	if commentFilter.SiteID != nil && *commentFilter.SiteID != "" {
		query += " AND site_id = ?"
		args = append(args, *commentFilter.SiteID)
	}

	if commentFilter.PostID != nil && *commentFilter.PostID != "" {
		query += " AND post_id = ?"
		args = append(args, *commentFilter.PostID)
	}

	if commentFilter.IsActive != nil {
		query += " AND is_active = ?"
		args = append(args, *commentFilter.IsActive)
	}

	empty := make([]conduit.Comment, 0)

	rows, err = cs.DB.Query(query, args...)
	if err != nil {
		logger.Error("query failed", "query", query, "args", args, "error", err)
		return empty, err
	}
	defer rows.Close()

	var comments []conduit.Comment

	for rows.Next() {
		var c conduit.Comment
		var t time.Time
		err = rows.Scan(&c.CommentID, &c.SiteID, &c.PostID, &t, &c.Author, &c.AuthorEmail, &c.CommentBody, &c.IsActive)
		if err != nil {
			logger.Error("scan failed", "error", err)
			return empty, err
		}
		c.Timestamp = conduit.Timestamp(t)

		comments = append(comments, c)
	}

	return comments, nil
}

func (cs *CommentService) DeleteComment(ctx context.Context, commentID string) error {
	logger := conduit.GetLogger(ctx)

	stmt, err := cs.DB.Prepare("DELETE FROM comments WHERE comment_id = ")
	if err != nil {
		logger.Error("failed to prepare DELETE query", "error", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(commentID)
	if err != nil {
		logger.Error("failed to DELETE comment", "error", err, "comment_id", commentID)
		return err
	}

	return nil
}
