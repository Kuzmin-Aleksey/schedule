package grpc_server

import (
	"context"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log/slog"
	"schedule/internal/domain/usecase/schedule"
	"schedule/internal/server/grpcserver"
	"schedule/internal/util"
	"schedule/pkg/contextx"
)

func NewGrpcServer(l *slog.Logger, schedule *schedule.Usecase) *grpc.Server {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandlerContext(func(ctx context.Context, p any) (err error) {
			l.ErrorContext(ctx, "recovered panic ", "err", p)
			return status.Error(codes.Internal, "internal error")
		}),
	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		addLoggerUnaryInterceptor(l),
		traceIdUnaryInterceptor,
		timezoneUnaryInterceptor,
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLog(l), loggingOpts...),
	))
	grpcserver.Register(server, schedule)
	return server
}

func addLoggerUnaryInterceptor(l *slog.Logger) func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = contextx.WithLogger(ctx, l)
		return handler(ctx, req)
	}
}

func timezoneUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tz := md.Get("TZ")
		if len(tz) > 0 {
			loc, err := util.ParseTimezone(tz[0])
			if err == nil {
				ctx = contextx.WithLocation(ctx, loc)
			}
		}
	}

	return handler(ctx, req)
}

const TraceIdMdKey = "X-Trace-Id"

func traceIdUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	var traceId string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		traceIdSlice := md.Get(TraceIdMdKey)
		if len(traceIdSlice) > 0 {
			if traceIdSlice[0] != "" {
				traceId = traceIdSlice[0]
			}
		}
	}

	if traceId == "" {
		traceId = uuid.NewString()
		header := metadata.Pairs(TraceIdMdKey, traceId)
		grpc.SetHeader(ctx, header)
	}

	ctx = contextx.WithTraceId(ctx, contextx.TraceId(traceId))

	return handler(ctx, req)
}
