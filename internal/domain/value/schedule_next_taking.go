package value

import "time"

type ScheduleNextTaking struct {
	time.Time
}

func (t ScheduleNextTaking) String() string {
	return t.Format(time.RFC3339)
}

func NewScheduleNextTaking(t time.Time) ScheduleNextTaking {
	return ScheduleNextTaking{
		Time: t,
	}
}
