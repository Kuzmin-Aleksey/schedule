package app

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"schedule/internal/app/logger"
	"schedule/internal/config"
	"schedule/internal/domain/usecase/schedule"
	"schedule/internal/infrastructure/persistence/mysql"
	"schedule/internal/server/grpcserver"
	"schedule/internal/server/httpserver"
	"schedule/pkg/contextx"
	"schedule/pkg/interceptorx"
	"schedule/pkg/middlwarex"
	"syscall"
)

func Run(cfg *config.Config) {
	l, err := logger.GetLogger(&cfg.Log)
	if err != nil {
		log.Fatal(err)
	}

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	db, err := mysql.Connect(cfg.MySQl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	scheduleRepo := mysql.NewScheduleRepo(db)

	scheduleUsecase := schedule.NewUsecase(scheduleRepo, cfg.Schedule)

	httpServer := newHttpServer(l, scheduleUsecase, cfg.HttpServer)
	grpcServer := newGrpcServer(l, scheduleUsecase)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	go func() {
		listener, err := net.Listen("tcp", cfg.GrpcServer.Addr)
		if err != nil {
			log.Fatal(err)
		}

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	<-shutdown

	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Println("shutdown http server failed:", err)
	}

	grpcServer.GracefulStop()
}

func newHttpServer(l *slog.Logger, schedule *schedule.Usecase, cfg config.HttpServerConfig) *http.Server {
	restScheduleServer := httpserver.NewScheduleServer(schedule)
	restServer := httpserver.NewServer(restScheduleServer)

	rtr := mux.NewRouter()
	restServer.RegisterRoutes(rtr)

	var sensitiveFields = []string{
		"user_id", "user-id", "userid",
	}

	rtr.Use(
		middlwarex.AddTraceId,
		middlwarex.WithLocation,
		middlwarex.NewLogRequest(&middlwarex.LogOptions{
			MaxContentLen:   cfg.Log.MaxRequestContentLen,
			LoggingContent:  cfg.Log.RequestLoggingContent,
			SensitiveFields: sensitiveFields,
		}),
		middlwarex.NewLogResponse(&middlwarex.LogOptions{
			MaxContentLen:   cfg.Log.MaxResponseContentLen,
			LoggingContent:  cfg.Log.ResponseLoggingContent,
			SensitiveFields: sensitiveFields,
		}),
	)

	return &http.Server{
		Handler:      rtr,
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorLog:     slog.NewLogLogger(l.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return contextx.WithLogger(context.Background(), l)
		},
	}
}

func newGrpcServer(l *slog.Logger, schedule *schedule.Usecase) *grpc.Server {
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

	safeField := []string{
		"user_id", "userid",
	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptorx.AddLoggerUnaryInterceptor(l),
		interceptorx.TraceIdUnaryInterceptor,
		interceptorx.TimezoneUnaryInterceptor,
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorx.NewLoggingInterceptor(safeField), loggingOpts...),
	))
	grpcserver.Register(server, schedule)
	return server
}
