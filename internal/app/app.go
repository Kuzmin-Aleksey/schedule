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
	"schedule/internal/app/logger"
	mysqlRepo "schedule/internal/repository/mysql"
	"schedule/internal/usecase/schedule"
	"syscall"
)

func Run(cfg *config.Config) {
	l, err := logger.GetLogger(&cfg.Log)
	if err != nil {
		log.Fatal(err)
	}

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	db, err := mysqlRepo.Connect(cfg.Db)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	scheduleRepo := mysqlRepo.NewScheduleRepo(db)
	if err := scheduleRepo.Migrate(); err != nil {
		log.Fatal("migrate failed: ", err)
	}

	scheduleUsecase := schedule.NewUsecase(scheduleRepo, l, cfg.Schedule)

	httpServer := newHttpServer(l, scheduleUsecase, cfg.HttpServer)
	grpcServer := NewGrpcServer(l, scheduleUsecase)

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
		l.Error("shutdown http server failed:", err)
	}

	grpcServer.GracefulStop()
}
