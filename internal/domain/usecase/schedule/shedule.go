package schedule

import (
	"context"
	"fmt"
	"log/slog"
	"schedule/internal/config"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/internal/util"
	"schedule/pkg/contextx"
	"time"
)

const day = 24 * time.Hour

type Repo interface {
	Save(ctx context.Context, schedule *entity.Schedule) error
	GetByUser(ctx context.Context, userId value.UserId) ([]entity.Schedule, error)
	GetById(ctx context.Context, userId value.UserId, scheduleId value.ScheduleId) (*entity.Schedule, error)
}

type Usecase struct {
	repo Repo
	l    *slog.Logger
	cfg  config.ScheduleConfig
}

func NewUsecase(repo Repo, l *slog.Logger, cfg config.ScheduleConfig) *Usecase {
	time.Local = nil
	return &Usecase{
		repo: repo,
		l:    l,
		cfg:  cfg,
	}
}

func (uc *Usecase) Create(ctx context.Context, dto *entity.ScheduleWithDuration) (value.ScheduleId, error) {
	var expiredAt *time.Time
	if dto.Duration > 0 {
		expiredAt = util.Ptr(time.Now().Add(time.Duration(dto.Duration) * day))
	}

	schedule := &entity.Schedule{
		UserId: dto.UserId,
		Name:   dto.Name,
		EndAt:  value.NewScheduleEndAt(expiredAt),
		Period: dto.Period,
	}

	if err := uc.repo.Save(ctx, schedule); err != nil {
		uc.l.ErrorContext(ctx, "create schedule error", "err", err)
		return 0, err
	}

	uc.l.DebugContext(ctx, "create schedule", "schedule", schedule)

	return schedule.Id, nil
}

func (uc *Usecase) GetByUser(ctx context.Context, userId value.UserId) ([]value.ScheduleId, error) {
	const op = "schedule.GetByUser"

	location := contextx.GetLocationOrDefault(ctx)

	schedules, err := uc.repo.GetByUser(ctx, userId)
	if err != nil {
		uc.l.ErrorContext(ctx, "get schedule by user error", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now().In(location)
	uc.l.DebugContext(ctx, op, "user time", now)

	ids := uc.getActualSchedulesIds(ctx, schedules, now, location)

	uc.l.DebugContext(ctx, op, "schedules", ids)

	return ids, nil
}

func (uc *Usecase) GetTimetable(ctx context.Context, userId value.UserId, scheduleId value.ScheduleId) (*entity.ScheduleTimetable, error) {
	const op = "schedule.GetTimetable"

	schedule, err := uc.repo.GetById(ctx, userId, scheduleId)
	if err != nil {
		uc.l.ErrorContext(ctx, "get schedule error", "err", err, "scheduleId", scheduleId)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// todo handle not found

	location := contextx.GetLocationOrDefault(ctx)

	uc.setScheduleEndHour(location, schedule)

	timetable := &entity.ScheduleTimetable{
		Id:        schedule.Id,
		Name:      schedule.Name,
		Period:    schedule.Period,
		EndAt:     schedule.EndAt,
		Timetable: []value.ScheduleTimeTableItem{},
	}

	now := time.Now().In(location)
	uc.l.DebugContext(ctx, op, "user time", now)

	if now.Round(time.Hour).Hour() > uc.cfg.EndDayHour { // if night then calculate for next day
		uc.l.DebugContext(ctx, "calculate for next day")
		now = now.Add(day)
	}

	if !schedule.EndAt.IsNil() && schedule.EndAt.Before(now) {
		uc.l.DebugContext(ctx, "schedule are expired", "schedule", schedule)
		return timetable, nil
	}

	timetable.Timetable = uc.makeTimetable(ctx, schedule, now, location)

	uc.l.DebugContext(ctx, op, "timetable", timetable)

	return timetable, nil
}

func (uc *Usecase) GetNextTakings(ctx context.Context, userId value.UserId) ([]entity.ScheduleNextTaking, error) {
	const op = "schedule.GetNextTakings"

	schedules, err := uc.repo.GetByUser(ctx, userId)
	if err != nil {
		uc.l.ErrorContext(ctx, "get schedule by user error", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	location := contextx.GetLocationOrDefault(ctx)
	now := time.Now().In(location)
	uc.l.DebugContext(ctx, op, "user time", now)

	nextTakings := uc.findNextTakings(ctx, schedules, now, location)

	uc.l.DebugContext(ctx, op, "NextTakings", nextTakings)

	return nextTakings, nil
}
