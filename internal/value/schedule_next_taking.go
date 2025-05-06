package value

import "time"

type ScheduleNextTaking struct {
	time.Time
}

func NewScheduleNextTaking(t time.Time) ScheduleNextTaking {
	return ScheduleNextTaking{
		Time: t,
	}
}
