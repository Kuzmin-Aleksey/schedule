package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/pkg/rest"
)

type ScheduleUsecase interface {
	Create(ctx context.Context, schedule *entity.ScheduleWithDuration) (value.ScheduleId, error)
	GetByUser(ctx context.Context, userId value.UserId) ([]value.ScheduleId, error)
	GetTimetable(ctx context.Context, userId value.UserId, scheduleId value.ScheduleId) (*entity.ScheduleTimetable, error)
	GetNextTakings(ctx context.Context, userId value.UserId) ([]entity.ScheduleNextTaking, error)
}

type ScheduleServer struct {
	Base
	schedule ScheduleUsecase
}

func NewScheduleServer(schedule ScheduleUsecase, base Base) ScheduleServer {
	return ScheduleServer{
		Base:     base,
		schedule: schedule,
	}
}

func (s *ScheduleServer) createSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := new(rest.CreateScheduleRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	schedule, err := newDomainScheduleWithDuration(req)
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	if err := schedule.Validate(); err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	id, err := s.schedule.Create(r.Context(), schedule)
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	s.writeJson(ctx, w, newRESTCreateScheduleResponse(id), http.StatusOK)
}

func (s *ScheduleServer) getUserSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	resp, err := s.schedule.GetByUser(ctx, userId)
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	s.writeJson(ctx, w, resp, http.StatusOK)
}

func (s *ScheduleServer) getSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}
	scheduleId, err := value.ParseScheduleId(r.FormValue("schedule_id"))
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	scheduleTimetable, err := s.schedule.GetTimetable(ctx, userId, scheduleId)
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	s.writeJson(ctx, w, newRESTScheduleResponse(scheduleTimetable), http.StatusOK)
}

func (s *ScheduleServer) scheduleGetNextTakings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := value.ParseUserId(r.FormValue("user_id"))
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	schedules, err := s.schedule.GetNextTakings(ctx, userId)
	if err != nil {
		s.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	s.writeJson(ctx, w, newRESTNextTakingResponse(schedules), http.StatusOK)
}
