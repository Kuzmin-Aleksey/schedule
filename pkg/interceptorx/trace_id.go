package interceptorx

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"schedule/pkg/contextx"
)

const traceIdMdKey = "X-Trace-Id"

func TraceIdUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	var traceId string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		traceIdSlice := md.Get(traceIdMdKey)
		if len(traceIdSlice) > 0 {
			if traceIdSlice[0] != "" {
				traceId = traceIdSlice[0]
			}
		}
	}

	if traceId == "" {
		traceId = uuid.NewString()
		header := metadata.Pairs(traceIdMdKey, traceId)
		grpc.SetHeader(ctx, header)
	}

	ctx = contextx.WithTraceId(ctx, contextx.TraceId(traceId))

	return handler(ctx, req)
}
