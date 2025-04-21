package app

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"schedule/config"
	"schedule/internal/controller/grpcHandler"
	"schedule/internal/usecase/schedule"
	"schedule/internal/util"
	"schedule/pkg/logger"
)

type GrpcServer struct {
	server *grpc.Server
	cfg    *config.Config
}

func NewGrpcServer(l *logger.Logger, schedule *schedule.Usecase) *grpc.Server {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			l.Error("recovered panic ", p)
			return status.Error(codes.Internal, "internal error")
		}),
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLog(l), loggingOpts...),
		timezoneUnaryInterceptor,
	))
	grpcHandler.Register(grpcServer, schedule, l)
	return grpcServer
}

func interceptorLog(l *logger.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		args := append([]any{msg}, fields...)

		switch level {
		case logging.LevelDebug, logging.LevelInfo:
			l.Debug(args)
		case logging.LevelWarn:
			l.Warn(args)
		case logging.LevelError:
			l.Error(args)
		}
	})
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
