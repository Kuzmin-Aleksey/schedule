package schedule

import (
	"context"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/internal/util"
	"schedule/pkg/contextx"
	"time"
)

func getActualSchedulesIds(ctx context.Context, schedules []*entity.Schedule) []value.ScheduleId {
	l := contextx.GetLoggerOrDefault(ctx)

	location := contextx.GetLocationOrDefault(ctx)
	now := time.Now().In(location)

	var ids []value.ScheduleId
	for _, schedule := range schedules {
		if schedule.EndAt.IsNil() || schedule.EndAt.After(now) {
			l.DebugContext(ctx, "add schedule", "schedule", schedule)
			ids = append(ids, schedule.Id)
		} else {
			l.DebugContext(ctx, "schedule expired", "schedule", schedule)
		}
	}

	return ids
}

func makeTimetable(ctx context.Context, schedule *entity.Schedule, beginDayHour, endDayHour int, round time.Duration) value.ScheduleTimeTable {
	l := contextx.GetLoggerOrDefault(ctx)

	location := contextx.GetLocationOrDefault(ctx)
	now := time.Now().In(location)

	timetable := value.ScheduleTimeTable{}

	beginOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), beginDayHour, 0, 0, 0, location)
	endOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), endDayHour, 0, 0, 0, location)

	for i := 0; ; i++ {
		timestamp := beginOfCurrentDay.Add(time.Duration(i) * time.Duration(schedule.Period))
		timestamp = timestamp.Round(round)

		if endOfCurrentDay.Before(timestamp) {
			l.DebugContext(ctx, "day end", "timestamp", timestamp)
			break
		}

		timetable = append(timetable, value.NewScheduleTimeTableItem(timestamp))
	}

	return timetable
}

func findNextTakings(ctx context.Context, schedules []*entity.Schedule, period time.Duration, beginDayHour, endDayHour int, round time.Duration) []entity.ScheduleNextTaking {
	l := contextx.GetLoggerOrDefault(ctx)

	location := contextx.GetLocationOrDefault(ctx)
	now := time.Now().In(location)

	nextTakingPeriod := now.Add(period)

	nextTakings := make([]entity.ScheduleNextTaking, 0) // if result is nil then write [] in json

	for _, schedule := range schedules {
		l.DebugContext(ctx, "finding taking", "schedule", schedule)

	DaysLoop:
		for days := 0; ; days++ {
			beginOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day()+days, beginDayHour, 0, 0, 0, location)
			l.DebugContext(ctx, "finding for day", "day", beginOfCurrentDay)

			for i := 0; ; i++ {
				timestamp := beginOfCurrentDay.Add(time.Duration(i) * time.Duration(schedule.Period))
				timestamp = timestamp.Round(round)
				l.DebugContext(ctx, "checking timestamp", "timestamp", timestamp)

				if !schedule.EndAt.IsNil() && timestamp.After(schedule.EndAt.ToTime()) { // if schedule end
					l.DebugContext(ctx, "schedule expired", "schedule", schedule, "timestamp", timestamp)
					break DaysLoop
				}

				if timestamp.After(nextTakingPeriod) {
					l.DebugContext(ctx, "schedule out of period", "schedule", schedule, "timestamp", timestamp)
					break DaysLoop
				}

				if timestamp.Hour() < beginDayHour || timestamp.Hour() >= endDayHour {
					l.DebugContext(ctx, "now night", "schedule", schedule, "timestamp", timestamp)
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

					l.DebugContext(ctx, "find next taking", "nextTaking", nextTaking)

					nextTakings = util.InsertFunc(nextTakings, nextTaking, func(v entity.ScheduleNextTaking) bool { // make sorted result
						return nextTaking.NextTaking.Before(v.NextTaking.Time)
					})
				}
			}
		}

	}

	return nextTakings
}
