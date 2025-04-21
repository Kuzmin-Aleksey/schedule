package httpHandler

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	_ "schedule/docs"
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

	h.rtr.Use(h.mwLogging)

	return h
}

func (h *Handler) SetScheduleRoutes(schedule *schedule.Usecase) {
	h.schedule = schedule

	h.rtr.HandleFunc("/schedule", h.createSchedule).Methods(http.MethodPost)
	h.rtr.HandleFunc("/schedule", h.mwWithLocation(h.getSchedule)).Methods(http.MethodGet)
	h.rtr.HandleFunc("/schedules", h.mwWithLocation(h.getUserSchedules)).Methods(http.MethodGet)
	h.rtr.HandleFunc("/next_taking", h.mwWithLocation(h.scheduleGetNextTakings)).Methods(http.MethodGet)
}

func (h *Handler) InitSwaggerHandler() {
	h.rtr.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rtr.ServeHTTP(w, r)
}
