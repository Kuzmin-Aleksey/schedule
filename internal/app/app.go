package app

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"schedule/config"
	mysqlRepo "schedule/internal/repository/mysql"
	"schedule/internal/usecase/schedule"
	"schedule/pkg/logger"
	"syscall"
)

func Run(cfg *config.Config) {
	l, err := logger.NewLogger(cfg.Log.File, cfg.Log.Level)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	db, err := mysqlRepo.Connect(cfg.Db)
	if err != nil {
		l.Fatal(err)
	}
	defer db.Close()

	scheduleRepo := mysqlRepo.NewScheduleRepo(db)
	if err := scheduleRepo.Migrate(); err != nil {
		l.Fatal("migrate failed: ", err)
	}

	scheduleUsecase := schedule.NewUsecase(scheduleRepo, cfg.Schedule)

	httpServer := newHttpServer(l, scheduleUsecase, cfg.HttpServer)
	grpcServer := NewGrpcServer(l, scheduleUsecase)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Fatal(err)
		}
	}()

	go func() {
		listener, err := net.Listen("tcp", cfg.GrpcServer.Addr)
		if err != nil {
			l.Fatal(err)
		}

		if err := grpcServer.Serve(listener); err != nil {
			l.Fatal(err)
		}
	}()

	<-shutdown

	if err := httpServer.Shutdown(context.Background()); err != nil {
		l.Error("shutdown http server failed:", err)
	}

	grpcServer.GracefulStop()
}
