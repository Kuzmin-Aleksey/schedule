package schedule

import (
	"context"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/internal/util"
	"time"
)

func (uc *Usecase) getActualSchedulesIds(ctx context.Context, schedules []entity.Schedule, now time.Time, location *time.Location) []value.ScheduleId {
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

	return ids
}

func (uc *Usecase) makeTimetable(ctx context.Context, schedule *entity.Schedule, now time.Time, location *time.Location) value.ScheduleTimeTable {
	timetable := value.ScheduleTimeTable{}

	beginOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), uc.cfg.BeginDayHour, 0, 0, 0, location)
	endOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), uc.cfg.EndDayHour, 0, 0, 0, location)

	for i := 0; ; i++ {
		timestamp := beginOfCurrentDay.Add(time.Duration(i) * time.Duration(schedule.Period))
		timestamp = timestamp.Round(uc.cfg.TimeRound)

		if endOfCurrentDay.Before(timestamp) {
			uc.l.DebugContext(ctx, "day end", "timestamp", timestamp)
			break
		}

		timetable = append(timetable, value.NewScheduleTimeTableItem(timestamp))
	}

	return timetable
}

func (uc *Usecase) findNextTakings(ctx context.Context, schedules []entity.Schedule, now time.Time, location *time.Location) []entity.ScheduleNextTaking {
	nextTakingPeriod := now.Add(uc.cfg.NextTakingPeriod)

	nextTakings := make([]entity.ScheduleNextTaking, 0) // if result is nil then write [] in json

	for _, schedule := range schedules {
		uc.setScheduleEndHour(location, &schedule)
		uc.l.DebugContext(ctx, "finding taking", "schedule", schedule)

	DaysLoop:
		for days := 0; ; days++ {
			beginOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day()+days, uc.cfg.BeginDayHour, 0, 0, 0, location)
			uc.l.DebugContext(ctx, "finding for day", "day", beginOfCurrentDay)

			for i := 0; ; i++ {
				timestamp := beginOfCurrentDay.Add(time.Duration(i) * time.Duration(schedule.Period))
				timestamp = timestamp.Round(uc.cfg.TimeRound)
				uc.l.DebugContext(ctx, "checking timestamp", "timestamp", timestamp)

				if !schedule.EndAt.IsNil() && timestamp.After(schedule.EndAt.ToTime()) { // if schedule end
					uc.l.DebugContext(ctx, "schedule expired", "schedule", schedule, "timestamp", timestamp)
					break DaysLoop
				}

				if timestamp.After(nextTakingPeriod) {
					uc.l.DebugContext(ctx, "schedule out of period", "schedule", schedule, "timestamp", timestamp)
					break DaysLoop
				}

				if timestamp.Hour() < uc.cfg.BeginDayHour || timestamp.Hour() >= uc.cfg.EndDayHour {
					uc.l.DebugContext(ctx, "now night", "schedule", schedule, "timestamp", timestamp)
					break
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

	}

	return nextTakings
}

func (uc *Usecase) setScheduleEndHour(loc *time.Location, s *entity.Schedule) { // in db this is DATE type without time
	if !s.EndAt.IsNil() {
		s.EndAt = value.NewScheduleEndAt(util.Ptr(time.Date(s.EndAt.Year(), s.EndAt.Month(), s.EndAt.Day(), uc.cfg.EndDayHour, 0, 0, 0, loc)))
	}
}
