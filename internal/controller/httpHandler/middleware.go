package httpHandler

import (
	"fmt"
	"net/http"
	"schedule/internal/usecase/schedule"
	"schedule/internal/util"
	"time"
)

type CustomWriter struct {
	http.ResponseWriter
	StatusCode    int
	ContentLength int
}

func (r *CustomWriter) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *CustomWriter) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)
	r.ContentLength += n
	return n, err
}

func (h *Handler) mwLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customWriter := &CustomWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(customWriter, r)
		end := time.Now()

		h.l.HttpLog(r.Method, r.URL.Path, customWriter.StatusCode, r.RemoteAddr, int(r.ContentLength), customWriter.ContentLength, end.Sub(start))
	})
}

func (h *Handler) mwWithLocation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locHeader := r.Header.Get("TZ")
		var loc *time.Location

		if locHeader != "" {
			var err error
			loc, err = util.ParseTimezone(locHeader)
			if err != nil {
				h.writeAndLogErr(w, fmt.Errorf("invalid timezone: %w", err), http.StatusBadRequest)
				return
			}
		} else {
			loc = time.UTC // default
		}

		//h.l.Debug("user time: ", time.Now().In(loc).Format("2006-01-02 15:04:05 -07:00"))

		ctx := schedule.CtxWithLocation(r.Context(), loc)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
