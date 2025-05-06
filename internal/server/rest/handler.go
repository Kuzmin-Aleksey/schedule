package rest

import (
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"schedule/config"
	"schedule/internal/usecase/schedule"
)

type Handler struct {
	rtr *mux.Router
	l   *slog.Logger

	maxLogContentReqLen  int64
	maxLogContentRespLen int
	logReqContent        map[string]struct{}
	logRespContent       map[string]struct{}

	schedule *schedule.Usecase
}

func NewHandler(l *slog.Logger, logCfg *config.HttpLog) *Handler {
	var h = &Handler{
		rtr:                  mux.NewRouter(),
		l:                    l,
		maxLogContentReqLen:  int64(logCfg.MaxResponseContentLen),
		maxLogContentRespLen: logCfg.MaxResponseContentLen,
		logReqContent:        make(map[string]struct{}),
		logRespContent:       make(map[string]struct{}),
	}

	for _, c := range logCfg.RequestLoggingContent {
		h.logReqContent[c] = struct{}{}
	}
	for _, c := range logCfg.ResponseLoggingContent {
		h.logRespContent[c] = struct{}{}
	}

	h.rtr.Use(
		h.mwAddTraceId,
		h.mwLogging,
	)

	return h
}

func (h *Handler) SetScheduleRoutes(schedule *schedule.Usecase) {
	h.schedule = schedule

	h.rtr.HandleFunc("/schedule", h.createSchedule).Methods(http.MethodPost)
	h.rtr.HandleFunc("/schedule", h.mwWithLocation(h.getSchedule)).Methods(http.MethodGet)
	h.rtr.HandleFunc("/schedules", h.mwWithLocation(h.getUserSchedules)).Methods(http.MethodGet)
	h.rtr.HandleFunc("/next_taking", h.mwWithLocation(h.scheduleGetNextTakings)).Methods(http.MethodGet)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rtr.ServeHTTP(w, r)
}
