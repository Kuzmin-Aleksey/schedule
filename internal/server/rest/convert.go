package rest

import (
	"schedule/internal/domain/entity"
	value2 "schedule/internal/domain/value"
	"schedule/internal/server/rest/models"
)

func newDomainScheduleWithDuration(req *models.CreateScheduleRequest) (*entity.ScheduleWithDuration, error) {
	period, err := value2.ParseSchedulePeriod(req.Period)
	if err != nil {
		return nil, err

	}

	return &entity.ScheduleWithDuration{
		UserId:   value2.UserId(req.UserId),
		Name:     value2.ScheduleName(req.Name),
		Duration: value2.ScheduleDuration(req.Duration),
		Period:   period,
	}, nil
}

func newRESTCreateScheduleResponse(id value2.ScheduleId) models.CreateScheduleResponse {
	return models.CreateScheduleResponse{
		Id: int(id),
	}
}

func newRESTScheduleResponse(timetable *entity.ScheduleTimetable) *models.ScheduleResponse {
	return &models.ScheduleResponse{
		Id:        int(timetable.Id),
		EndAt:     timetable.EndAt.NullableString(),
		Name:      string(timetable.Name),
		Period:    timetable.Period.String(),
		Timetable: timetable.Timetable.ToStringArray(),
	}
}

func newRESTNextTakingResponse(schedules []entity.ScheduleNextTaking) []*models.NextTakingResponse {
	resp := make([]*models.NextTakingResponse, len(schedules))

	for i, t := range schedules {
		resp[i] = &models.NextTakingResponse{
			Id:         int(t.Id),
			EndAt:      t.EndAt.NullableString(),
			Name:       string(t.Name),
			NextTaking: t.NextTaking.String(),
			Period:     t.Period.String(),
		}
	}

	return resp
}
