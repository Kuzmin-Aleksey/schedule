package interceptorx

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"schedule/internal/util"
	"schedule/pkg/contextx"
)

const timezoneMDKey = "TZ"

func TimezoneUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tz := md.Get(timezoneMDKey)
		if len(tz) > 0 {
			loc, err := util.ParseTimezone(tz[0])
			if err == nil {
				ctx = contextx.WithLocation(ctx, loc)
			}
		}
	}

	return handler(ctx, req)
}
