package schedule

import (
	"context"
	"fmt"
	"log/slog"
	"schedule/config"
	"schedule/internal/entity"
	"schedule/internal/util"
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
	GetByUser(ctx context.Context, userId int64) ([]entity.Schedule, error)
	GetById(ctx context.Context, userId int64, scheduleId int) (*entity.Schedule, error)
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

func (uc *Usecase) Create(ctx context.Context, dto *CreateScheduleDTO) (*CreateScheduleResponseDTO, error) {
	var expiredAt *time.Time
	if dto.Duration > 0 {
		expiredAt = util.Ptr(time.Now().Add(time.Duration(dto.Duration) * day))
	}

	schedule := &entity.Schedule{
		UserId: dto.UserId,
		Name:   dto.Name,
		EndAt:  expiredAt,
		Period: dto.Period,
	}

	if err := uc.repo.Save(ctx, schedule); err != nil {
		uc.l.ErrorContext(ctx, "create schedule error", "err", err)
		return nil, err
	}

	uc.l.DebugContext(ctx, "create schedule", "schedule", schedule)

	return &CreateScheduleResponseDTO{
		Id: schedule.Id,
	}, nil
}

func (uc *Usecase) GetByUser(ctx context.Context, userId int64) ([]int, error) {
	const op = "schedule.GetByUser"

	location := getLocationCtx(ctx)

	schedules, err := uc.repo.GetByUser(ctx, userId)
	if err != nil {
		uc.l.ErrorContext(ctx, "get schedule by user error", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now().In(location)
	uc.l.DebugContext(ctx, op, "user time", now)

	var ids []int
	for _, schedule := range schedules {
		uc.setScheduleEndHour(location, &schedule)
		if schedule.EndAt == nil || schedule.EndAt.After(now) {
			ids = append(ids, schedule.Id)
		} else {
			uc.l.DebugContext(ctx, "schedule expired", "schedule", schedule)
		}
	}

	uc.l.DebugContext(ctx, op, "schedules", ids)

	return ids, nil
}

func (uc *Usecase) GetTimetable(ctx context.Context, userId int64, scheduleId int) (*ScheduleResponseDTO, error) {
	const op = "schedule.GetTimetable"

	schedule, err := uc.repo.GetById(ctx, userId, scheduleId)
	if err != nil {
		uc.l.ErrorContext(ctx, "get schedule error", "err", err, "scheduleId", scheduleId)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	location := getLocationCtx(ctx)

	uc.setScheduleEndHour(location, schedule)

	dto := &ScheduleResponseDTO{
		Id:        schedule.Id,
		Name:      schedule.Name,
		Period:    schedule.Period,
		EndAt:     schedule.EndAt,
		Timetable: []time.Time{},
	}

	now := time.Now().In(location)
	uc.l.DebugContext(ctx, op, "user time", now)

	if now.Round(time.Hour).Hour() > uc.cfg.EndDayHour { // if night then calculate for next day
		uc.l.DebugContext(ctx, "calculate for next day")
		now = now.Add(day)
	}

	if schedule.EndAt != nil && schedule.EndAt.Before(now) {
		return dto, nil
	}

	startOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), uc.cfg.BeginDayHour, 0, 0, 0, location)
	endOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), uc.cfg.EndDayHour, 0, 0, 0, location)

	for i := 0; ; i++ {
		timestamp := startOfCurrentDay.Add(time.Duration(i) * schedule.Period)
		timestamp = timestamp.Round(uc.cfg.TimeRound)

		if endOfCurrentDay.Before(timestamp) {
			break
		}

		dto.Timetable = append(dto.Timetable, timestamp)
	}

	uc.l.DebugContext(ctx, op, "timetable", dto)

	return dto, nil
}

func (uc *Usecase) GetNextTakings(ctx context.Context, userId int64) ([]NextTakingResponseDTO, error) {
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

	dto := make([]NextTakingResponseDTO, 0) // if result is nil then write [] in json

	beginOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), uc.cfg.BeginDayHour, 0, 0, 0, location)

	for _, schedule := range schedules {
		uc.setScheduleEndHour(location, &schedule)
		uc.l.DebugContext(ctx, "finding taking", "schedule", schedule)

		for i := 0; ; i++ {
			timestamp := beginOfCurrentDay.Add(time.Duration(i) * schedule.Period)
			timestamp = timestamp.Round(uc.cfg.TimeRound)

			if schedule.EndAt != nil && timestamp.After(*schedule.EndAt) { // if schedule end
				uc.l.DebugContext(ctx, "schedule expired", "schedule", schedule)
				break
			}

			if timestamp.After(nextTakingPeriod) {
				uc.l.DebugContext(ctx, "schedule out of period", "schedule", schedule)
				break
			}

			if timestamp.After(now) && timestamp.Hour() >= uc.cfg.BeginDayHour && timestamp.Hour() < uc.cfg.EndDayHour {
				nextTaking := NextTakingResponseDTO{
					Id:         schedule.Id,
					Name:       schedule.Name,
					EndAt:      schedule.EndAt,
					Period:     schedule.Period,
					NextTaking: timestamp,
				}

				uc.l.DebugContext(ctx, "find next taking", "nextTaking", nextTaking)

				dto = util.InsertFunc(dto, nextTaking, func(v NextTakingResponseDTO) bool { // make sorted result
					return nextTaking.NextTaking.Before(v.NextTaking)
				})

				break
			}
		}
	}

	uc.l.DebugContext(ctx, op, "NextTakings", dto)

	return dto, nil
}

func (uc *Usecase) setScheduleEndHour(loc *time.Location, s *entity.Schedule) { // in db this is DATE type without time
	if s.EndAt != nil {
		*s.EndAt = (*s.EndAt).Add(time.Duration(uc.cfg.EndDayHour) * time.Hour).In(loc)
	}
}
