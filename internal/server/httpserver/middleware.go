package httpserver

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"schedule/internal/util"
	"schedule/pkg/contextx"
	"time"
)

func (s *Server) mwWithLocation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locHeader := r.Header.Get("TZ")
		var loc *time.Location

		if locHeader != "" {
			var err error
			loc, err = util.ParseTimezone(locHeader)
			if err != nil {
				s.writeAndLogErr(r.Context(), w, fmt.Errorf("invalid timezone: %w", err), http.StatusBadRequest)
				return
			}
		} else {
			loc = time.UTC // default
		}

		ctx := contextx.WithLocation(r.Context(), loc)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

const headerTraceId = "X-Trace-Id"

func (s *Server) mwAddTraceId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := r.Header.Get(headerTraceId)
		if traceId == "" {
			traceId = uuid.NewString()
			w.Header().Set(headerTraceId, traceId)
		}

		ctx := contextx.WithTraceId(r.Context(), contextx.TraceId(traceId))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
