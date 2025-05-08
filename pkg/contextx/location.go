package contextx

import (
	"context"
	"time"
)

type contextKeyLocation struct{}

func WithLocation(ctx context.Context, location *time.Location) context.Context {
	return context.WithValue(ctx, contextKeyLocation{}, location)
}

var DefaultLocation = time.UTC

func GetLocationOrDefault(ctx context.Context) *time.Location {
	loc, ok := ctx.Value(contextKeyLocation{}).(*time.Location)

	if !ok || loc == nil {
		return DefaultLocation
	}

	return loc
}
