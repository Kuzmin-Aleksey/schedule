package server

import (
	"context"
	"schedule/internal/domain/aggregate"
	"schedule/internal/domain/value"
)

type ScheduleUsecase interface {
	Create(ctx context.Context, schedule *aggregate.ScheduleWithDuration) (value.ScheduleId, error)
	GetByUser(ctx context.Context, userId value.UserId) ([]value.ScheduleId, error)
	GetTimetable(ctx context.Context, userId value.UserId, scheduleId value.ScheduleId) (*aggregate.ScheduleWithTimetable, error)
	GetNextTakings(ctx context.Context, userId value.UserId) ([]aggregate.ScheduleNextTaking, error)
}
