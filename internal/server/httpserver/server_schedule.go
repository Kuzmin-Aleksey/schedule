package httpserver

import (
	"encoding/json"
	"net/http"
	"schedule/internal/domain/value"
	"schedule/internal/server"
	"schedule/pkg/failure"
	"schedule/pkg/rest"
)

type ScheduleServer struct {
	schedule server.ScheduleUsecase
}

func NewScheduleServer(schedule server.ScheduleUsecase) ScheduleServer {
	return ScheduleServer{
		schedule: schedule,
	}
}

func (s *ScheduleServer) createSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := new(rest.CreateScheduleRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	schedule, err := newDomainScheduleWithDuration(req)
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	if err := schedule.Validate(); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	id, err := s.schedule.Create(r.Context(), schedule)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, newRESTCreateScheduleResponse(id), http.StatusOK)
}

func (s *ScheduleServer) getUserSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	resp, err := s.schedule.GetByUser(ctx, userId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, resp, http.StatusOK)
}

func (s *ScheduleServer) getSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}
	scheduleId, err := value.ParseScheduleId(r.FormValue("schedule_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	scheduleTimetable, err := s.schedule.GetTimetable(ctx, userId, scheduleId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, newRESTScheduleResponse(scheduleTimetable), http.StatusOK)
}

func (s *ScheduleServer) scheduleGetNextTakings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	schedules, err := s.schedule.GetNextTakings(ctx, userId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, newRESTNextTakingResponse(schedules), http.StatusOK)
}
