package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/carlohamalainen/carlo-comments/conduit"
	"github.com/google/uuid"
)

func getClientIP(r *http.Request) string {
    ip := r.Header.Get("X-Real-IP")
    if ip != "" {
        return ip
    }

    ips := r.Header.Get("X-Forwarded-For")
    if ips != "" {
        ipList := strings.Split(ips, ",")
        if len(ipList) > 0 {
            return strings.TrimSpace(ipList[0])
        }
    }

    return r.RemoteAddr
}

func logRequestHeaders(logger *slog.Logger, r *http.Request) {
    headers := make(map[string]string)
    for name, values := range r.Header {
		switch len(values) {
		case 0:
			continue
		case 1:
			headers[name] = values[0]
		default:
			headers[name] = fmt.Sprintf("%v", values)
		}
    }

    logger.Info("Request headers",
        "method", r.Method,
        "url", r.URL.String(),
        "headers", headers,
    )
}

func (s *Server) createComment() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.Logger.With("request_id", uuid.NewString(), "handler", "createComment", "client_ip", getClientIP(r))
		ctx := conduit.WithLogger(r.Context(), logger)

		logRequestHeaders(logger, r)

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

		if newComment.TurnstileToken == "" {
			logger.Error("turnstile token empty")
			http.Error(w, "Internal server error", http.StatusBadRequest)
			return
		}

		if _, err := VerifyTurnstileToken(newComment.TurnstileToken, s.Config.CfSecretKey); err != nil {
			logger.Error("turnstile token rejected", "error", err, "turnstile_token", newComment.TurnstileToken)
			http.Error(w, "Internal server error", http.StatusBadRequest)
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

		comment.SourceAddress = getClientIP(r)
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

		go func() {
			err := Notify(&s.Config, &comment)
			if err != nil {
				logger.Error("failed to send notification email", "error", err.Error())
			} else {
				logger.Info("sent notification email", "to", s.Config.AdminUser)

			}
		}()

		w.WriteHeader(http.StatusCreated)
	}
}

func (s *Server) getComments(redact bool, filterMode FilterMode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.Logger.With("request_id", uuid.NewString(), "handler", "getComments")
		ctx := conduit.WithLogger(r.Context(), logger)

		if r.Method != http.MethodPost {
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

		if redact {
			for i := range comments {
				comments[i].AuthorEmail = ""
				comments[i].SourceAddress = ""
			}
		}
		writeJSON(ctx, w, http.StatusOK, comments)
	}
}

func (s *Server) upsertComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.Logger.With("request_id", uuid.NewString(), "handler", "upsertComment", "source_ip", r.RemoteAddr)
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
