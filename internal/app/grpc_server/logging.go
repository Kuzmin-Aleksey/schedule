package grpc_server

import (
	"context"
	"fmt"
	"github.com/brunoga/deep"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"log/slog"
	"reflect"
	"slices"
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
				if err := hideSafeValues(c); err != nil {
					l.ErrorContext(ctx, "hide safe fields failed", "err", err)
				}
				fields[i+1] = c
			}
		}
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

var safeFields = []string{
	"user_id",
	"userid",
}

func hideSafeValues(s any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

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
				if e := hideSafeValues(v); e != nil {
					err = e
				}
			}
		}

	case reflect.Slice:
		for i := 0; i < ps.Len(); i++ {
			if e := hideSafeValues(ps.Index(i)); e != nil {
				err = e
			}
		}
	case reflect.Map:
		for _, k := range ps.MapKeys() {
			if slices.Contains(safeFields, strings.ToLower(k.String())) {
				ps.SetMapIndex(k, reflect.Zero(ps.MapIndex(k).Type()))
			} else {
				if e := hideSafeValues(ps.MapIndex(k)); e != nil {
					err = e
				}
			}
		}
	default:
		return
	}

	return
}
