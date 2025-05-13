package interceptorx

import (
	"context"
	"fmt"
	"github.com/brunoga/deep"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"log/slog"
	"reflect"
	"schedule/pkg/contextx"
	"slices"
	"strings"
)

func NewLoggingInterceptor(safeFields []string) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		l := contextx.GetLoggerOrDefault(ctx)

		for i, f := range fields {
			if f == "grpc.request.content" || f == "grpc.response.content" {
				c, err := deep.CopySkipUnsupported(fields[i+1])
				if err != nil {
					l.ErrorContext(ctx, "copy message failed", "err", err)
					break
				}
				if err := hideSafeValues(c, safeFields); err != nil {
					l.ErrorContext(ctx, "hide safe fields failed", "err", err)
				}
				fields[i+1] = c
			}
		}
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

func hideSafeValues(s any, safeFields []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if s == nil {
		return nil
	}

	var ps reflect.Value

	if v, ok := s.(reflect.Value); ok {
		ps = v
	} else {
		ps = reflect.ValueOf(s)
	}

	if ps.Kind() != reflect.Ptr {
		if !ps.CanAddr() {
			return
		}
		ps = ps.Addr()
	}

	if ps.IsNil() {
		return nil
	}

	ps = ps.Elem()

	t := ps.Type()

	switch t.Kind() {
	case reflect.Struct:
		for i := range t.NumField() {
			f := t.Field(i)
			v := ps.Field(i)

			if slices.Contains(safeFields, strings.ToLower(f.Name)) && v.CanSet() {
				v.Set(reflect.Zero(f.Type))
			} else {
				if e := hideSafeValues(v, safeFields); e != nil {
					err = e
				}
			}
		}

	case reflect.Slice:
		for i := 0; i < ps.Len(); i++ {
			if e := hideSafeValues(ps.Index(i), safeFields); e != nil {
				err = e
			}
		}
	case reflect.Map:
		for _, k := range ps.MapKeys() {
			if slices.Contains(safeFields, strings.ToLower(k.String())) {
				ps.SetMapIndex(k, reflect.Zero(ps.MapIndex(k).Type()))
			} else {
				if e := hideSafeValues(ps.MapIndex(k), safeFields); e != nil {
					err = e
				}
			}
		}
	default:
		return
	}

	return
}
