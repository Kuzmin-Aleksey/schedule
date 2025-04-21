package app

import (
	"log"
	"net/http"
	"schedule/config"
	"schedule/internal/controller/httpHandler"
	"schedule/internal/usecase/schedule"
	"schedule/pkg/logger"
)

func newHttpServer(l *logger.Logger, schedule *schedule.Usecase, cfg config.HttpServerConfig) *http.Server {
	handler := httpHandler.NewHandler(l)
	handler.SetScheduleRoutes(schedule)
	handler.InitSwaggerHandler()

	return &http.Server{
		Handler:      handler,
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorLog:     log.New(l.Out, "", log.Ldate|log.Ltime),
	}
}
