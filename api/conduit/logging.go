package conduit

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/carlohamalainen/carlo-comments/config"
)

type ctxKey string

const loggerKey ctxKey = "slog-logger"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		panic("context is missing logger")
	}
	return logger
}

func NewLogger(cfg config.Config) (func(), *slog.Logger) {
	handler := newDailyFileHandler(cfg.LogDirectory, cfg.AppName)
	logger := slog.New(handler)
	return handler.Flush, logger
}

type dailyFileHandler struct {
	logDir    string
	logPrefix string
	attrs     map[string]slog.Value
}

func newDailyFileHandler(logDir, logPrefix string) *dailyFileHandler {
	return &dailyFileHandler{
		logDir:    logDir,
		logPrefix: logPrefix,
		attrs:     make(map[string]slog.Value),
	}
}

func (h *dailyFileHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *dailyFileHandler) Handle(_ context.Context, r slog.Record) error {
	logFile := h.getLogFileName()
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	logEntry := make(map[string]interface{})

	for k, v := range h.attrs {
		logEntry[k] = v.Any()
	}

	logEntry["timestamp"] = r.Time.Format(time.RFC3339)
	logEntry["level"] = r.Level.String()
	logEntry["message"] = r.Message

	// These can override the existing attributes set with "With()".
	r.Attrs(func(attr slog.Attr) bool {
		logEntry[attr.Key] = attr.Value.Any()
		return true // keep iterating through attributes
	})

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}
	_, err = file.WriteString("\n")
	return err
}

func (h *dailyFileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	for _, attr := range attrs {
		h.attrs[attr.Key] = attr.Value
	}
	return h
}

func (h *dailyFileHandler) WithGroup(name string) slog.Handler {
	panic("WithGroup not implemented")
}

func (h *dailyFileHandler) Flush() {
	// TODO What might we do here?
}

func (h *dailyFileHandler) getLogFileName() string {
	currentTime := time.Now()
	logFileName := h.logPrefix + "-" + currentTime.Format("2006-01-02") + ".log"
	return filepath.Join(h.logDir, logFileName)
}
