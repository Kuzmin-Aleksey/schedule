package server

import (
	"context"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
)

type ScheduleUsecase interface {
	Create(ctx context.Context, schedule *entity.ScheduleWithDuration) (value.ScheduleId, error)
	GetByUser(ctx context.Context, userId value.UserId) ([]value.ScheduleId, error)
	GetTimetable(ctx context.Context, userId value.UserId, scheduleId value.ScheduleId) (*entity.ScheduleTimetable, error)
	GetNextTakings(ctx context.Context, userId value.UserId) ([]entity.ScheduleNextTaking, error)
}
