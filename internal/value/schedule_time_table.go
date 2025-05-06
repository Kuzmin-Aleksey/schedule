package value

import "time"

type ScheduleTimeTableItem struct {
	time.Time
}

func NewScheduleTimeTableItem(t time.Time) ScheduleTimeTableItem {
	return ScheduleTimeTableItem{
		Time: t,
	}
}

type ScheduleTimeTable []ScheduleTimeTableItem

func (t ScheduleTimeTable) ToStringArray() []string {
	s := make([]string, len(t))
	for i, item := range t {
		s[i] = item.String()
	}
	return s
}
