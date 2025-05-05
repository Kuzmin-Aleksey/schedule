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
	"schedule/config"
	"schedule/internal/app/logger"
	"schedule/internal/controller/grpcHandler"
	"schedule/internal/usecase/schedule"
	"schedule/internal/util"
)

type GrpcServer struct {
	server *grpc.Server
	cfg    *config.Config
}

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

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		traceIdUnaryInterceptor,
		timezoneUnaryInterceptor,
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLog(l), loggingOpts...),
	))
	grpcHandler.Register(grpcServer, schedule, l)
	return grpcServer
}

func timezoneUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tz := md.Get("TZ")
		if len(tz) > 0 {
			loc, err := util.ParseTimezone(tz[0])
			if err == nil {
				ctx = schedule.CtxWithLocation(ctx, loc)
			}
		}
	}

	return handler(ctx, req)
}

func traceIdUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	var traceId string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		traceIdSlice := md.Get("X-Trace-Id")
		if len(traceIdSlice) > 0 {
			if traceIdSlice[0] != "" {
				traceId = traceIdSlice[0]
			}
		}
	}

	if traceId == "" {
		traceId = uuid.NewString()
		header := metadata.Pairs("X-Trace-Id", traceId)
		grpc.SetHeader(ctx, header)
	}

	ctx = context.WithValue(ctx, logger.TraceIdKey{}, traceId)

	return handler(ctx, req)
}
