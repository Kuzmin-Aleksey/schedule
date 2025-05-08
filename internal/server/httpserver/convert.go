package httpserver

import (
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/pkg/rest"
)

func newDomainScheduleWithDuration(req *rest.CreateScheduleRequest) (*entity.ScheduleWithDuration, error) {
	period, err := value.ParseSchedulePeriod(req.Period)
	if err != nil {
		return nil, err

	}

	return &entity.ScheduleWithDuration{
		UserId:   value.UserId(req.UserId),
		Name:     value.ScheduleName(req.Name),
		Duration: value.ScheduleDuration(req.Duration),
		Period:   period,
	}, nil
}

func newRESTCreateScheduleResponse(id value.ScheduleId) rest.CreateScheduleResponse {
	return rest.CreateScheduleResponse{
		Id: int(id),
	}
}

func newRESTScheduleResponse(timetable *entity.ScheduleTimetable) *rest.ScheduleResponse {
	return &rest.ScheduleResponse{
		Id:        int(timetable.Id),
		EndAt:     timetable.EndAt.NullableString(),
		Name:      string(timetable.Name),
		Period:    timetable.Period.String(),
		Timetable: timetable.Timetable.ToStringArray(),
	}
}

func newRESTNextTakingResponse(schedules []entity.ScheduleNextTaking) []*rest.NextTakingResponse {
	resp := make([]*rest.NextTakingResponse, len(schedules))

	for i, t := range schedules {
		resp[i] = &rest.NextTakingResponse{
			Id:         int(t.Id),
			EndAt:      t.EndAt.NullableString(),
			Name:       string(t.Name),
			NextTaking: t.NextTaking.String(),
			Period:     t.Period.String(),
		}
	}

	return resp
}
