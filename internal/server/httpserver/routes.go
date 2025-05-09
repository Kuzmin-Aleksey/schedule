package httpserver

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) RegisterRoutes(rtr *mux.Router) {
	rtr.HandleFunc("/schedule", s.createSchedule).Methods(http.MethodPost)
	rtr.HandleFunc("/schedule", s.getSchedule).Methods(http.MethodGet)
	rtr.HandleFunc("/schedules", s.getUserSchedules).Methods(http.MethodGet)
	rtr.HandleFunc("/next_taking", s.scheduleGetNextTakings).Methods(http.MethodGet)
}
