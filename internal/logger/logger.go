package logger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Arzeeq/pvz-api/internal/config"
	"github.com/Arzeeq/pvz-api/internal/dto"
)

const (
	LogFormatText = "text"
	LogFormatJson = "json"
)

type MyLogger struct {
	*slog.Logger
}

func New(env string, format string) *MyLogger {
	var loggerLevel slog.Level

	switch env {
	case config.EnvTest:
		loggerLevel = slog.LevelDebug
	case config.EnvDev:
		loggerLevel = slog.LevelDebug
	case config.EnvProd:
		loggerLevel = slog.LevelInfo
	}

	var logger *slog.Logger
	switch format {
	case LogFormatJson:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
	case LogFormatText:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
	default:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
		logger.Warn(fmt.Sprintf("unsupported logging format %s, using default format instead", format))
	}

	return &MyLogger{logger}
}

func (l *MyLogger) WrapError(msg string, err error, args ...any) {
	args = append(args, slog.String("error", err.Error()))
	l.Error(msg, args...)
}

func (l *MyLogger) HTTPResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	errJSONEncode := json.NewEncoder(w).Encode(payload)
	if errJSONEncode != nil {
		l.WrapError("failed to write response", errJSONEncode)
	}
}

func (l *MyLogger) HTTPError(w http.ResponseWriter, status int, err error) {
	l.HTTPResponse(w, status, dto.Error{Message: err.Error()})
}
