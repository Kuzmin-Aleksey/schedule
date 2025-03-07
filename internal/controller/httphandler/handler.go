package httphandler

import (
	"github.com/gorilla/mux"
	"net/http"
	"schedule/internal/usecase/schedule"
	"schedule/pkg/logger"
)

type Handler struct {
	rtr *mux.Router
	l   *logger.Logger

	schedule *schedule.Usecase
}

func NewHandler(l *logger.Logger) *Handler {
	var h = &Handler{
		rtr: mux.NewRouter(),
		l:   l,
	}

	h.rtr.Use(h.MwLogging)

	return h
}

func (h *Handler) SetScheduleRoutes(schedule *schedule.Usecase) {
	h.schedule = schedule

	h.rtr.HandleFunc("/schedule", h.createSchedule).Methods("POST")
	h.rtr.HandleFunc("/schedule", h.MwWithLocation(h.getSchedule)).Methods("GET")
	h.rtr.HandleFunc("/schedules", h.MwWithLocation(h.getUserSchedules)).Methods("GET")
	h.rtr.HandleFunc("/next_taking", h.MwWithLocation(h.scheduleGetNextTakings)).Methods("GET")
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rtr.ServeHTTP(w, r)
}
