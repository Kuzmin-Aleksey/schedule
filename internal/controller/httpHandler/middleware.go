package httpHandler

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"schedule/internal/app/logger"
	"schedule/internal/usecase/schedule"
	"schedule/internal/util"
	"time"
)

func (h *Handler) mwWithLocation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locHeader := r.Header.Get("TZ")
		var loc *time.Location

		if locHeader != "" {
			var err error
			loc, err = util.ParseTimezone(locHeader)
			if err != nil {
				h.writeAndLogErr(r.Context(), w, fmt.Errorf("invalid timezone: %w", err), http.StatusBadRequest)
				return
			}
		} else {
			loc = time.UTC // default
		}

		ctx := schedule.CtxWithLocation(r.Context(), loc)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *Handler) MwAddTraceId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := r.Header.Get("X-Trace-Id")
		if traceId == "" {
			traceId = uuid.NewString()
			w.Header().Set("X-Trace-Id", traceId)
		}

		ctx := context.WithValue(r.Context(), logger.TraceIdKey{}, traceId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
