package schedule

import (
	"context"
	"fmt"
	"log/slog"
	"schedule/config"
	"schedule/internal/entity"
	"schedule/internal/util"
	"schedule/internal/value"
	"time"
)

const day = 24 * time.Hour

type userLocationCtxKey struct{}

func CtxWithLocation(ctx context.Context, location *time.Location) context.Context {
	return context.WithValue(ctx, userLocationCtxKey{}, location)
}

func getLocationCtx(ctx context.Context) *time.Location {
	location := time.UTC
	if loc, ok := ctx.Value(userLocationCtxKey{}).(*time.Location); loc != nil && ok {
		location = loc
	}
	return location
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=Repo
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

	location := getLocationCtx(ctx)

	schedules, err := uc.repo.GetByUser(ctx, userId)
	if err != nil {
		uc.l.ErrorContext(ctx, "get schedule by user error", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now().In(location)
	uc.l.DebugContext(ctx, op, "user time", now)

	var ids []value.ScheduleId
	for _, schedule := range schedules {
		uc.setScheduleEndHour(location, &schedule)
		if schedule.EndAt.IsNil() || schedule.EndAt.After(now) {
			uc.l.DebugContext(ctx, "add schedule", "schedule", schedule)
			ids = append(ids, schedule.Id)
		} else {
			uc.l.DebugContext(ctx, "schedule expired", "schedule", schedule)
		}
	}

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

	location := getLocationCtx(ctx)

	uc.setScheduleEndHour(location, schedule)

	dto := &entity.ScheduleTimetable{
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
		return dto, nil
	}

	beginOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), uc.cfg.BeginDayHour, 0, 0, 0, location)
	endOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), uc.cfg.EndDayHour, 0, 0, 0, location)

	for i := 0; ; i++ {
		timestamp := beginOfCurrentDay.Add(time.Duration(i) * time.Duration(schedule.Period))
		timestamp = timestamp.Round(uc.cfg.TimeRound)

		if endOfCurrentDay.Before(timestamp) {
			break
		}

		dto.Timetable = append(dto.Timetable, value.NewScheduleTimeTableItem(timestamp))
	}

	uc.l.DebugContext(ctx, op, "timetable", dto)

	return dto, nil
}

func (uc *Usecase) GetNextTakings(ctx context.Context, userId value.UserId) ([]entity.ScheduleNextTaking, error) {
	const op = "schedule.GetNextTakings"

	schedules, err := uc.repo.GetByUser(ctx, userId)
	if err != nil {
		uc.l.ErrorContext(ctx, "get schedule by user error", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	location := getLocationCtx(ctx)
	now := time.Now().In(location)
	uc.l.DebugContext(ctx, op, "user time", now)

	nextTakingPeriod := now.Add(uc.cfg.NextTakingPeriod)

	nextTakings := make([]entity.ScheduleNextTaking, 0) // if result is nil then write [] in json

	beginOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), uc.cfg.BeginDayHour, 0, 0, 0, location)

	for _, schedule := range schedules {
		uc.setScheduleEndHour(location, &schedule)
		uc.l.DebugContext(ctx, "finding taking", "schedule", schedule)

		for i := 0; ; i++ {
			timestamp := beginOfCurrentDay.Add(time.Duration(i) * time.Duration(schedule.Period))
			timestamp = timestamp.Round(uc.cfg.TimeRound)
			uc.l.DebugContext(ctx, "checking timestamp", "timestamp", timestamp)

			if !schedule.EndAt.IsNil() && timestamp.After(schedule.EndAt.ToTime()) { // if schedule end
				uc.l.DebugContext(ctx, "schedule expired", "schedule", schedule, "timestamp", timestamp)
				break
			}

			if timestamp.After(nextTakingPeriod) {
				uc.l.DebugContext(ctx, "schedule out of period", "schedule", schedule, "timestamp", timestamp)
				break
			}

			if timestamp.Hour() < uc.cfg.BeginDayHour || timestamp.Hour() >= uc.cfg.EndDayHour {
				uc.l.DebugContext(ctx, "now night", "schedule", schedule, "timestamp", timestamp)
				continue
			}

			if timestamp.After(now) {
				nextTaking := entity.ScheduleNextTaking{
					Id:         schedule.Id,
					Name:       schedule.Name,
					EndAt:      schedule.EndAt,
					Period:     schedule.Period,
					NextTaking: value.NewScheduleNextTaking(timestamp),
				}

				uc.l.DebugContext(ctx, "find next taking", "nextTaking", nextTaking)

				nextTakings = util.InsertFunc(nextTakings, nextTaking, func(v entity.ScheduleNextTaking) bool { // make sorted result
					return nextTaking.NextTaking.Before(v.NextTaking.Time)
				})
			}
		}
	}

	uc.l.DebugContext(ctx, op, "NextTakings", nextTakings)

	return nextTakings, nil
}

func (uc *Usecase) setScheduleEndHour(loc *time.Location, s *entity.Schedule) { // in db this is DATE type without time
	if !s.EndAt.IsNil() {
		s.EndAt = value.NewScheduleEndAt(util.Ptr(time.Date(s.EndAt.Year(), s.EndAt.Month(), s.EndAt.Day(), uc.cfg.EndDayHour, 0, 0, 0, loc)))
	}
}
