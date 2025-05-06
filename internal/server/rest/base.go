package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"schedule/internal/server/rest/models"
)

type Base struct {
	l *slog.Logger
}

func NewBase(l *slog.Logger) Base {
	return Base{
		l: l,
	}
}

const errEncodingJson = "json encode error"

func (b *Base) writeAndLogErr(ctx context.Context, w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()}); err != nil {
		b.l.LogAttrs(ctx, slog.LevelError, errEncodingJson, slog.String("err", err.Error()))
	}

	b.l.LogAttrs(ctx, slog.LevelError, "error handling request", slog.String("err", err.Error()))
}

func (b *Base) writeJson(ctx context.Context, w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		b.l.LogAttrs(ctx, slog.LevelError, errEncodingJson, slog.String("err", err.Error()))
	}
}
