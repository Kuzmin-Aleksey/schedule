package httphandler

import (
	"encoding/json"
	"net/http"
	"schedule/internal/usecase/schedule"
	"strconv"
)

func (h *Handler) createSchedule(w http.ResponseWriter, r *http.Request) {
	dto := new(schedule.CreateScheduleDTO)
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}
	if err := dto.Validate(); err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.Create(r.Context(), dto)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(w, resp, http.StatusOK)
}

func (h *Handler) getUserSchedules(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.GetByUser(r.Context(), userId)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(w, resp, http.StatusOK)
}

func (h *Handler) getSchedule(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}
	scheduleId, err := strconv.Atoi(r.FormValue("schedule_id"))
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.GetTimetable(r.Context(), userId, scheduleId)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(w, resp, http.StatusOK)
}

func (h *Handler) scheduleGetNextTakings(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.GetNextTakings(r.Context(), userId)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(w, resp, http.StatusOK)
}
