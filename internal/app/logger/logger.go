package logger

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"schedule/internal/config"
	"schedule/pkg/contextx"
	"strings"
)

const (
	LogFormatJson = "json"
	LogFormatText = "text"
)

type logHandler struct {
	slog.Handler
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	traceId := contextx.GetTraceId(ctx)

	if traceId != "" {
		r.AddAttrs(slog.String("trace_id", string(traceId)))
	}

	return h.Handler.Handle(ctx, r)
}

func GetLogger(cfg *config.LogConfig) (*slog.Logger, error) {
	out := io.Writer(os.Stdout)

	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			return nil, err
		}
		out = io.MultiWriter(file, os.Stdout)
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case LogFormatText:
		handler = slog.NewTextHandler(out, opts)
	case LogFormatJson:
		handler = slog.NewJSONHandler(out, opts)
	default:
		log.Println("unknown logging format ", cfg.Format)
		handler = slog.NewJSONHandler(out, nil)
	}

	l := slog.New(&logHandler{
		Handler: handler,
	})

	return l, nil

}
