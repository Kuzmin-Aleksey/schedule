package interceptorx

import (
	"context"
	"google.golang.org/grpc"
	"log/slog"
	"schedule/pkg/contextx"
)

func AddLoggerUnaryInterceptor(l *slog.Logger) func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = contextx.WithLogger(ctx, l)
		return handler(ctx, req)
	}
}
