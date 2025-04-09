package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"schedule/config"
	"schedule/internal/controller/httphandler"
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

	handler := httphandler.NewHandler(l)
	handler.SetScheduleRoutes(scheduleUsecase)

	server := &http.Server{
		Handler:      handler,
		Addr:         cfg.HttpServer.Addr,
		ReadTimeout:  cfg.HttpServer.ReadTimeout,
		WriteTimeout: cfg.HttpServer.WriteTimeout,
		ErrorLog:     log.New(l.Out, "", log.Ldate|log.Ltime),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Fatal(err)
		}
	}()

	<-shutdown

	if err := server.Shutdown(context.Background()); err != nil {
		l.Error("shutdown server failed:", err)
	}
}
