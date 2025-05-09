package middlwarex

import (
	"log/slog"
	"net/http"
	"schedule/internal/util"
	"schedule/pkg/contextx"
	"time"
)

const timeZoneHeaderKey = "TZ"

func WithLocation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		locHeader := r.Header.Get(timeZoneHeaderKey)
		var loc *time.Location

		if locHeader != "" {
			var err error
			loc, err = util.ParseTimezone(locHeader)
			if err != nil {
				l := contextx.GetLoggerOrDefault(ctx)
				l.WarnContext(ctx, "failed parsing timezone", slog.String("err", err.Error()), slog.String("timezone", locHeader))
			}
		}

		if loc == nil {
			loc = time.Local
		}

		ctx = contextx.WithLocation(r.Context(), loc)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
