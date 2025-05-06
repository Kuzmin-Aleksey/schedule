package app

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"schedule/internal/app/grpc_server"
	"schedule/internal/app/logger"
	"schedule/internal/config"
	"schedule/internal/domain/usecase/schedule"
	"schedule/internal/infrastructure/persistence/mysql"
	"schedule/internal/server/rest"
	"syscall"
)

func Run(cfg *config.Config) {
	l, err := logger.GetLogger(&cfg.Log)
	if err != nil {
		log.Fatal(err)
	}

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	db, err := mysql.Connect(cfg.Db)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	scheduleRepo := mysql.NewScheduleRepo(db)
	if err := scheduleRepo.Migrate(); err != nil {
		log.Fatal("migrate failed: ", err)
	}

	scheduleUsecase := schedule.NewUsecase(scheduleRepo, l, cfg.Schedule)

	httpServer := newHttpServer(l, scheduleUsecase, cfg.HttpServer)
	grpcServer := grpc_server.NewGrpcServer(l, scheduleUsecase)

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
	base := rest.NewBase(l)
	restScheduleServer := rest.NewScheduleServer(schedule, base)
	restServer := rest.NewServer(restScheduleServer, l, &cfg.Log)

	rtr := mux.NewRouter()
	restServer.RegisterRoutes(rtr)

	return &http.Server{
		Handler:      rtr,
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorLog:     slog.NewLogLogger(l.Handler(), slog.LevelError),
	}
}
