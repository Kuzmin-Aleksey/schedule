package httpHandler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"schedule/internal/controller/httpHandler/models"
)

const errEncodingJson = "json encode error"

func (h *Handler) writeAndLogErr(ctx context.Context, w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()}); err != nil {
		h.l.LogAttrs(ctx, slog.LevelError, errEncodingJson, slog.String("err", err.Error()))
	}

	h.l.LogAttrs(ctx, slog.LevelError, "error handling request", slog.String("err", err.Error()))

	w.WriteHeader(status)
}

func (h *Handler) writeJson(ctx context.Context, w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.l.LogAttrs(ctx, slog.LevelError, errEncodingJson, slog.String("err", err.Error()))
	}
}
