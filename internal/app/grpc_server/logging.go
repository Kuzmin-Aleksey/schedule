package grpc_server

import (
	"context"
	"github.com/brunoga/deep"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"log/slog"
	"reflect"
	"strings"
)

func interceptorLog(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		for i, f := range fields {
			if f == "grpc.request.content" || f == "grpc.response.content" {
				c, err := deep.CopySkipUnsupported(fields[i+1])
				if err != nil {
					l.ErrorContext(ctx, "copy message failed", "err", err)
					break
				}
				hideSafeValues(c)
				fields[i+1] = c
			}
		}
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

var safeFields = map[string]struct{}{
	"user_id": {},
	"userid":  {},
}

func hideSafeValues(s any) {
	var ps reflect.Value

	if v, ok := s.(reflect.Value); ok {
		ps = v
	} else {
		ps = reflect.ValueOf(s)
	}

	if ps.Kind() != reflect.Ptr {
		ps = ps.Addr()
	}
	ps = ps.Elem()

	t := ps.Type()

	for i := range t.NumField() {
		f := t.Field(i)
		v := ps.Field(i)

		if !v.CanInterface() {
			continue
		}

		switch f.Type.Kind() {
		case reflect.Struct:
			hideSafeValues(v)
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				hideSafeValues(v.Index(i))
			}
		case reflect.Map:
			for _, k := range v.MapKeys() {
				hideSafeValues(v.MapIndex(k))
			}
		default:
			if _, ok := safeFields[strings.ToLower(f.Name)]; ok && v.CanSet() {
				v.Set(reflect.Zero(f.Type))
			}
		}
	}
}
