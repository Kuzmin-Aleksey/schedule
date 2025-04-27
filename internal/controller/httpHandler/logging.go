package httpHandler

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type LoggingWriter struct {
	http.ResponseWriter
	StatusCode    int
	ContentLength int
	Content       []byte
	MaxContentLen int
}

func (r *LoggingWriter) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *LoggingWriter) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)
	r.ContentLength += n

	if r.ContentLength > r.MaxContentLen {
		r.Content = nil
	} else {
		r.Content = append(r.Content, p...)
	}

	return n, err
}

func (h *Handler) mwLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logRequest(r)

		customWriter := &LoggingWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
			MaxContentLen:  h.maxLogContentRespLen,
		}

		start := time.Now()
		next.ServeHTTP(customWriter, r)
		end := time.Now()

		h.logResponse(r.Context(), customWriter, end.Sub(start))
	})
}

func (h *Handler) logRequest(r *http.Request) {
	ctx := r.Context()

	attrs := []slog.Attr{
		slog.String("protocol", r.Proto),
		slog.String("method", r.Method),
		slog.String("url", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("content_length", r.ContentLength),
		slog.String("content_type", r.Header.Get("Content-Type")),
	}

	if r.Form == nil {
		r.ParseForm()
	}

	if len(r.Form) > 0 {
		attrs = append(attrs, slog.Group("values", getSafeSlogValues(r.Form)...))
	}

	if _, logContent := h.logReqContent[r.Header.Get("Content-Type")]; r.ContentLength > 0 && r.ContentLength <= h.maxLogContentReqLen && logContent {
		content, err := io.ReadAll(r.Body)
		if err != nil {
			content = []byte{}
			h.l.LogAttrs(ctx, slog.LevelError, "read request body error", slog.String("err", err.Error()))
		}
		attrs = append(attrs, slog.String("content", string(content)))
		r.Body = io.NopCloser(bytes.NewReader(content))
	}

	h.l.LogAttrs(ctx, slog.LevelInfo, "request received", attrs...)
}

func getSafeSlogValues(v url.Values) []any {
	if v == nil {
		return nil
	}

	attrs := make([]any, 0, len(v))

	for key := range v {
		var val string
		if key == "user_id" {
			for range len(v.Get(key)) {
				val += "*"
			}
		} else {
			val = v.Get(key)
		}

		attrs = append(attrs, slog.String(key, val))
	}

	return attrs
}

func (h *Handler) logResponse(ctx context.Context, r *LoggingWriter, handleDuration time.Duration) {
	contentType := r.Header().Get("Content-Type")

	attrs := []slog.Attr{
		slog.String("status", strconv.Itoa(r.StatusCode)),
		slog.Duration("handle_duration", handleDuration),
		slog.String("content_type", contentType),
		slog.Int("content_len", r.ContentLength),
	}

	if _, logContent := h.logRespContent[contentType]; len(r.Content) > 0 && logContent {
		attrs = append(attrs, slog.String("content", string(r.Content)))
	}

	h.l.LogAttrs(ctx, slog.LevelInfo, "response sent", attrs...)
}
