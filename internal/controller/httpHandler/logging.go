package httpHandler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"schedule/internal/util"
	"slices"
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

	if r.ContentLength+n < r.MaxContentLen {
		r.Content = append(r.Content, p...)
	} else if r.ContentLength < r.MaxContentLen {
		r.Content = append(r.Content, p[:r.MaxContentLen-r.ContentLength]...)
	}

	r.ContentLength += n

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

	contentType := r.Header.Get("Content-Type")

	attrs := []slog.Attr{
		slog.String("protocol", r.Proto),
		slog.String("method", r.Method),
		slog.String("url", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("content_length", r.ContentLength),
		slog.String("content_type", contentType),
	}

	if r.Form == nil {
		r.ParseForm()
	}

	if len(r.Form) > 0 {
		attrs = append(attrs, slog.Group("values", getSafeSlogValues(r.Form)...))
	}

	if _, logContent := h.logReqContent[contentType]; r.ContentLength > 0 && logContent {
		body := make([]byte, min(r.ContentLength, h.maxLogContentReqLen))
		if _, err := r.Body.Read(body); err != nil && !errors.Is(err, io.EOF) {
			h.l.LogAttrs(ctx, slog.LevelError, "read request body error", slog.String("err", err.Error()))
		}

		if contentType == "application/json" {
			unmarshalledBody := util.JsonUnmarshal(body)
			hideSafeValues(unmarshalledBody)

			attrs = append(attrs, slog.Any("content", unmarshalledBody))
		} else {
			attrs = append(attrs, slog.String("content", string(body)))
		}

		r.Body = util.NewMultiReadCloser(io.NopCloser(bytes.NewReader(body)), r.Body)
	}

	h.l.LogAttrs(ctx, slog.LevelInfo, "request received", attrs...)
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
		if contentType == "application/json" {
			unmarshalledBody := util.JsonUnmarshal(r.Content)
			hideSafeValues(unmarshalledBody)

			attrs = append(attrs, slog.Any("content", unmarshalledBody))
		} else {
			attrs = append(attrs, slog.String("content", string(r.Content)))
		}
	}

	h.l.LogAttrs(ctx, slog.LevelInfo, "response sent", attrs...)
}

var safeFields = []string{
	"user_id",
	"userid",
	"user-id",
}

func hideSafeValues(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k := range t {
			if slices.Contains(safeFields, k) {
				t[k] = "hidden"
			} else {
				hideSafeValues(t[k])
			}
		}
	case []any:
		for i := range t {
			hideSafeValues(t[i])
		}
	}
}

func getSafeSlogValues(v url.Values) []any {
	if v == nil {
		return nil
	}

	attrs := make([]any, 0, len(v))

	for k := range v {
		var val string
		if slices.Contains(safeFields, k) {
			val = "hidden"
		} else {
			val = v.Get(k)
		}

		attrs = append(attrs, slog.String(k, val))
	}

	return attrs
}
