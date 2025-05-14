package aggregate

import "schedule/internal/domain/value"

type ScheduleNextTaking struct {
	Id         value.ScheduleId
	Name       value.ScheduleName
	EndAt      value.ScheduleEndAt
	Period     value.SchedulePeriod
	NextTaking value.ScheduleNextTaking
}
