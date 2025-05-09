package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"schedule/pkg/contextx"
	"schedule/pkg/errcodes"
	"schedule/pkg/failure"
	"schedule/pkg/rest"
)

func writeAndLogErr(ctx context.Context, w http.ResponseWriter, err error) {
	errCode, statusCode := getCodeFromError(err)
	writeJson(ctx, w, rest.ErrorResponse{Error: errCode.String()}, statusCode)

	l := contextx.GetLoggerOrDefault(ctx)
	l.LogAttrs(ctx, slog.LevelError, "error handling request", slog.String("err", err.Error()))
}

func writeJson(ctx context.Context, w http.ResponseWriter, v any, status int) {
	l := contextx.GetLoggerOrDefault(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		l.LogAttrs(ctx, slog.LevelError, "json encode error", slog.String("err", err.Error()))
	}
}

func getCodeFromError(err error) (errcodes.Code, int) {
	switch {
	case failure.IsInternalError(err):
		return errcodes.Internal, http.StatusInternalServerError
	case failure.IsNotFoundError(err):
		return errcodes.NotFound, http.StatusNotFound
	case failure.IsInvalidRequestError(err):
		return errcodes.Validation, http.StatusBadRequest
	default:
		return errcodes.Internal, http.StatusInternalServerError
	}
}
