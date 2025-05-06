package rest

import (
	"encoding/json"
	"net/http"
	value2 "schedule/internal/domain/value"
	"schedule/internal/server/rest/models"
)

func (h *Handler) createSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := new(models.CreateScheduleRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	schedule, err := newDomainScheduleWithDuration(req)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	if err := schedule.Validate(); err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	id, err := h.schedule.Create(r.Context(), schedule)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(ctx, w, newRESTCreateScheduleResponse(id), http.StatusOK)
}

func (h *Handler) getUserSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value2.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.GetByUser(ctx, userId)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(ctx, w, resp, http.StatusOK)
}

func (h *Handler) getSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value2.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}
	scheduleId, err := value2.ParseScheduleId(r.FormValue("schedule_id"))
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	scheduleTimetable, err := h.schedule.GetTimetable(ctx, userId, scheduleId)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(ctx, w, newRESTScheduleResponse(scheduleTimetable), http.StatusOK)
}

func (h *Handler) scheduleGetNextTakings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value2.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	schedules, err := h.schedule.GetNextTakings(ctx, userId)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(ctx, w, newRESTNextTakingResponse(schedules), http.StatusOK)
}
