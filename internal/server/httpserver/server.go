package httpserver

import (
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"schedule/internal/config"
)

type Server struct {
	ScheduleServer
	l *slog.Logger

	cfg *config.HttpLog
}

func NewServer(scheduleServer ScheduleServer, l *slog.Logger, logCfg *config.HttpLog) *Server {
	var h = &Server{
		ScheduleServer: scheduleServer,
		l:              l,
		cfg:            logCfg,
	}

	return h
}

func (s *Server) RegisterRoutes(rtr *mux.Router) {
	rtr.Use(
		s.mwAddTraceId,
		s.mwLogging,
	)

	rtr.HandleFunc("/schedule", s.createSchedule).Methods(http.MethodPost)
	rtr.HandleFunc("/schedule", s.mwWithLocation(s.getSchedule)).Methods(http.MethodGet)
	rtr.HandleFunc("/schedules", s.mwWithLocation(s.getUserSchedules)).Methods(http.MethodGet)
	rtr.HandleFunc("/next_taking", s.mwWithLocation(s.scheduleGetNextTakings)).Methods(http.MethodGet)
}
