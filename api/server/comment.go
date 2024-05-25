package server

import (
	"net/http"
	"time"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/google/uuid"
)

func (s *Server) createComment() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.Logger.With("request_id", uuid.NewString(), "handler", "createComment")
		ctx := conduit.WithLogger(r.Context(), logger)

		if r.Method != http.MethodPost {
			// TODO add to conduit/errors.go
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var newComment conduit.NewComment

		if err := readJSON(ctx, r.Body, &newComment, s.Config.MaxBodySize); err != nil {
			logger.Error("json decode failed", "error", err)
			badRequestError(ctx, w)
			return
		}

		nr, err := s.commentService.NrComments(ctx, conduit.CommentFilter{SiteID: &newComment.SiteID, PostID: &newComment.PostID})
		if err != nil {
			logger.Error("failed to count nr comments on post", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if nr >= s.Config.MaxNrComments {
			logger.Info("discarding comment due to over capcity", "site_id", newComment.SiteID, "post_id", newComment.PostID)
			http.Error(w, "Internal server error", http.StatusForbidden)
			return
		}

		comment := conduit.Comment{
			CommentID: uuid.NewString(),
			Timestamp: conduit.Timestamp(time.Now()),
			IsActive:  false,
		}

		if !conduit.IsValidPostID(newComment.PostID) {
			// TODO add to conduit/errors.go
			http.Error(w, "Invalid key", http.StatusBadRequest)
			return
		}
		comment.PostID = newComment.PostID

		if !s.IsKnown(newComment.SiteID, newComment.PostID) {
			// TODO add to conduit/errors.go
			logger.Error("unknown siteID and postID", "site_id", newComment.SiteID, "post_id", newComment.PostID)
			http.Error(w, "Unknown host", http.StatusBadRequest)
			return
		}
		comment.SiteID = newComment.SiteID

		comment.Author = Sanitize(newComment.Author)
		comment.CommentBody = Sanitize(newComment.CommentBody)

		if conduit.IsValidEmail(newComment.AuthorEmail) {
			comment.AuthorEmail = newComment.AuthorEmail
		}

		err = s.commentService.UpsertComment(ctx, &comment)
		if err != nil {
			// TODO add to conduit/errors.go
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (s *Server) getComments(filterMode FilterMode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.Logger.With("request_id", uuid.NewString(), "handler", "getComments")
		ctx := conduit.WithLogger(r.Context(), logger)

		if r.Method != http.MethodGet {
			// TODO add to conduit/errors.go
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		maxLength := s.Config.MaxBodySize
		if filterMode == FreeRange {
			maxLength = 0
		}

		var commentFilter conduit.CommentFilter
		if err := readJSON(ctx, r.Body, &commentFilter, maxLength); err != nil {
			logger.Error("failed to decode", "error", err)
			badRequestError(ctx, w)
			return
		}

		switch filterMode {
		case ActiveOnly:
			t := true
			commentFilter.IsActive = &t

			if commentFilter.SiteID == nil {
				// TODO add to conduit/errors.go
				logger.Error("active query is missing SiteID")
				http.Error(w, "missing SiteID in filter", http.StatusBadRequest)
				return
			}

			if commentFilter.PostID == nil {
				// TODO add to conduit/errors.go
				logger.Error("active query is missing PostID")
				http.Error(w, "missing PostID in filter", http.StatusBadRequest)
				return
			}

		case FreeRange:
			// no changes to the supplied filter
		default:
			// TODO add to conduit/errors.go
			logger.Error("unknown query filter", "filter_mode", filterMode)
			http.Error(w, "unknown filter mode", http.StatusInternalServerError)
			return
		}

		comments, err := s.commentService.Comments(ctx, commentFilter)
		if err != nil {
			// TODO add to conduit/errors.go
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		writeJSON(ctx, w, http.StatusOK, comments)
	}
}

func (s *Server) upsertComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.Logger.With("request_id", uuid.NewString(), "handler", "upsertComment")
		ctx := conduit.WithLogger(r.Context(), logger)

		if r.Method != http.MethodPost {
			// TODO add to conduit/errors.go
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var comment conduit.Comment

		if err := readJSON(ctx, r.Body, &comment, 0); err != nil {
			// TODO add to conduit/errors.go
			logger.Error("failed to decode json", "error", err)
			badRequestError(ctx, w)
			return
		}

		if !conduit.IsValidPostID(comment.PostID) {
			// TODO add to conduit/errors.go
			logger.Error("invalid postID", "post_id", comment.PostID)
			http.Error(w, "Invalid key", http.StatusBadRequest)
			return
		}

		if !s.IsKnown(comment.SiteID, comment.PostID) {
			// TODO add to conduit/errors.go
			logger.Error("unknown SiteID-PostID", "post_id", comment.PostID, "site_id", comment.SiteID)
			http.Error(w, "Unknown host", http.StatusBadRequest)
			return
		}

		comment.Author = Sanitize(comment.Author)
		comment.CommentBody = Sanitize(comment.CommentBody)

		if !conduit.IsValidEmail(comment.AuthorEmail) {
			// TODO add to conduit/errors.go
			logger.Error("invalid author email", "email", comment.AuthorEmail)
			http.Error(w, "bad email", http.StatusBadRequest)
			return
		}

		err := s.commentService.UpsertComment(ctx, &comment)
		if err != nil {
			// TODO add to conduit/errors.go
			logger.Error("upsert comment failed", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
