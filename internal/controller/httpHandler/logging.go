package httpHandler

import (
	"bytes"
	"context"
	"encoding/json"
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

	if _, logContent := h.logReqContent[contentType]; r.ContentLength > 0 && r.ContentLength <= h.maxLogContentReqLen && logContent {
		content, err := io.ReadAll(r.Body)
		if err != nil {
			content = []byte{}
			h.l.LogAttrs(ctx, slog.LevelError, "read request body error", slog.String("err", err.Error()))
		}

		if contentType == "application/json" {
			var mapContent map[string]interface{}

			if err := json.Unmarshal(content, &mapContent); err != nil {
				h.l.ErrorContext(ctx, "unmarshal request body to map error", slog.String("err", err.Error()))
			}

			hideSafeValuesInMap(mapContent)

			attrs = append(attrs, slog.Any("content", mapContent))
		} else {
			attrs = append(attrs, slog.String("content", string(content)))
		}

		r.Body = io.NopCloser(bytes.NewReader(content))
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
			var mapContent map[string]interface{}

			if err := json.Unmarshal(r.Content, &mapContent); err != nil {
				h.l.ErrorContext(ctx, "unmarshal response body to map error", slog.String("err", err.Error()))
			}

			hideSafeValuesInMap(mapContent)

			attrs = append(attrs, slog.Any("content", mapContent))
		} else {
			attrs = append(attrs, slog.String("content", string(r.Content)))
		}
	}

	h.l.LogAttrs(ctx, slog.LevelInfo, "response sent", attrs...)
}

var safeFields = map[string]struct{}{
	"user_id": {},
	"userid":  {},
	"user-id": {},
}

func hideSafeValuesInMap(m map[string]any) {
	for k := range m {
		if _, ok := m[k].(map[string]any); ok {
			hideSafeValuesInMap(m[k].(map[string]any))
		}

		if _, ok := safeFields[k]; ok {
			m[k] = "hidden"
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
		if _, ok := safeFields[k]; ok {
			val = "hidden"
		} else {
			val = v.Get(k)
		}

		attrs = append(attrs, slog.String(k, val))
	}

	return attrs
}
