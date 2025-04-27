package app

import (
	"log/slog"
	"net/http"
	"schedule/config"
	"schedule/internal/controller/httpHandler"
	"schedule/internal/usecase/schedule"
)

func newHttpServer(l *slog.Logger, schedule *schedule.Usecase, cfg config.HttpServerConfig) *http.Server {
	handler := httpHandler.NewHandler(l, &cfg.Log)
	handler.SetScheduleRoutes(schedule)
	handler.InitSwaggerHandler()

	return &http.Server{
		Handler:      handler,
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorLog:     slog.NewLogLogger(l.Handler(), slog.LevelError),
	}
}
