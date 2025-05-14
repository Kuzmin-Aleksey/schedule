package aggregate

import "schedule/internal/domain/value"

type ScheduleWithTimetable struct {
	Id        value.ScheduleId
	Name      value.ScheduleName
	EndAt     value.ScheduleEndAt
	Period    value.SchedulePeriod
	Timetable value.ScheduleTimeTable
}
