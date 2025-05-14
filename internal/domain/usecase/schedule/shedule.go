package schedule

import (
	"context"
	"fmt"
	"schedule/internal/config"
	"schedule/internal/domain/aggregate"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/internal/util"
	"schedule/pkg/contextx"
	"time"
)

const day = 24 * time.Hour

type Repo interface {
	Save(ctx context.Context, schedule *entity.Schedule) error
	GetByUser(ctx context.Context, userId value.UserId) ([]*entity.Schedule, error)
	GetById(ctx context.Context, userId value.UserId, scheduleId value.ScheduleId) (*entity.Schedule, error)
}

type Usecase struct {
	repo Repo
	cfg  config.ScheduleConfig
}

func NewUsecase(repo Repo, cfg config.ScheduleConfig) *Usecase {
	time.Local = nil
	return &Usecase{
		repo: repo,
		cfg:  cfg,
	}
}

func (uc *Usecase) Create(ctx context.Context, dto *aggregate.ScheduleWithDuration) (value.ScheduleId, error) {
	const op = "schedule.Create"

	l := contextx.GetLoggerOrDefault(ctx)

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
		l.ErrorContext(ctx, "create schedule error", "err", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	l.DebugContext(ctx, "create schedule", "schedule", schedule)

	return schedule.Id, nil
}

func (uc *Usecase) GetByUser(ctx context.Context, userId value.UserId) ([]value.ScheduleId, error) {
	const op = "schedule.GetByUser"

	l := contextx.GetLoggerOrDefault(ctx)

	location := contextx.GetLocationOrDefault(ctx)

	schedules, err := uc.repo.GetByUser(ctx, userId)
	if err != nil {
		l.ErrorContext(ctx, "get schedule by user error", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	uc.setScheduleEndHour(location, schedules)

	now := time.Now().In(location)
	l.DebugContext(ctx, op, "user time", now)

	ids := getActualSchedulesIds(ctx, schedules)

	l.DebugContext(ctx, op, "schedules", ids)

	return ids, nil
}

func (uc *Usecase) GetTimetable(ctx context.Context, userId value.UserId, scheduleId value.ScheduleId) (*aggregate.ScheduleWithTimetable, error) {
	const op = "schedule.GetTimetable"

	l := contextx.GetLoggerOrDefault(ctx)

	schedule, err := uc.repo.GetById(ctx, userId, scheduleId)
	if err != nil {
		l.ErrorContext(ctx, "get schedule error", "err", err, "scheduleId", scheduleId)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	location := contextx.GetLocationOrDefault(ctx)

	uc.setScheduleEndHour(location, []*entity.Schedule{schedule})

	timetable := &aggregate.ScheduleWithTimetable{
		Id:        schedule.Id,
		Name:      schedule.Name,
		Period:    schedule.Period,
		EndAt:     schedule.EndAt,
		Timetable: []value.ScheduleTimeTableItem{},
	}

	now := time.Now().In(location)
	l.DebugContext(ctx, op, "user time", now)

	if now.Round(time.Hour).Hour() > uc.cfg.EndDayHour { // if night then calculate for next day
		l.DebugContext(ctx, "calculate for next day")
		now = now.Add(day)
	}

	if !schedule.EndAt.IsNil() && schedule.EndAt.Before(now) {
		l.DebugContext(ctx, "schedule are expired", "schedule", schedule)
		return timetable, nil
	}

	timetable.Timetable = makeTimetable(ctx, schedule, uc.cfg.BeginDayHour, uc.cfg.EndDayHour, uc.cfg.TimeRound)

	l.DebugContext(ctx, op, "timetable", timetable)

	return timetable, nil
}

func (uc *Usecase) GetNextTakings(ctx context.Context, userId value.UserId) ([]aggregate.ScheduleNextTaking, error) {
	const op = "schedule.GetNextTakings"

	l := contextx.GetLoggerOrDefault(ctx)

	schedules, err := uc.repo.GetByUser(ctx, userId)
	if err != nil {
		l.ErrorContext(ctx, "get schedule by user error", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	location := contextx.GetLocationOrDefault(ctx)
	now := time.Now().In(location)
	l.DebugContext(ctx, op, "user time", now)

	uc.setScheduleEndHour(location, schedules)

	nextTakings := findNextTakings(ctx, schedules, uc.cfg.NextTakingPeriod, uc.cfg.BeginDayHour, uc.cfg.EndDayHour, uc.cfg.TimeRound)

	l.DebugContext(ctx, op, "NextTakings", nextTakings)

	return nextTakings, nil
}

func (uc *Usecase) setScheduleEndHour(loc *time.Location, schedules []*entity.Schedule) { // in db this is DATE type without time
	for _, s := range schedules {
		if !s.EndAt.IsNil() {
			s.EndAt = value.NewScheduleEndAt(util.Ptr(time.Date(s.EndAt.Year(), s.EndAt.Month(), s.EndAt.Day(), uc.cfg.EndDayHour, 0, 0, 0, loc)))
		}
	}
}
