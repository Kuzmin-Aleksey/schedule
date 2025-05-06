package entity

import (
	"errors"
	value2 "schedule/internal/domain/value"
	"time"
)

const (
	MaxMedicineNameLen = 255
	MinSchedulePeriod  = value2.SchedulePeriod(time.Hour)
	MaxSchedulePeriod  = value2.SchedulePeriod(time.Hour * 24)
)

type Schedule struct {
	Id     value2.ScheduleId     `db:"id"`
	UserId value2.UserId         `db:"user_id" json:"-"` // med police 16 digits, always int64
	Name   value2.ScheduleName   `db:"name"`
	EndAt  value2.ScheduleEndAt  `db:"end_at"`
	Period value2.SchedulePeriod `db:"period"`
}

type ScheduleWithDuration struct {
	Id       value2.ScheduleId
	UserId   value2.UserId
	Name     value2.ScheduleName
	Duration value2.ScheduleDuration
	Period   value2.SchedulePeriod
}

func (t ScheduleWithDuration) Validate() error {
	switch {
	case t.UserId == 0:
		return errors.New("user id is required")
	case t.Name == "":
		return errors.New("name is required")
	case len(t.Name) > MaxMedicineNameLen:
		return errors.New("medicine name is too long")
	case t.Period < MinSchedulePeriod:
		return errors.New("period is too short")
	case t.Period > MaxSchedulePeriod:
		return errors.New("period is too long")
	}
	return nil
}

type ScheduleNextTaking struct {
	Id         value2.ScheduleId
	Name       value2.ScheduleName
	EndAt      value2.ScheduleEndAt
	Period     value2.SchedulePeriod
	NextTaking value2.ScheduleNextTaking
}

type ScheduleTimetable struct {
	Id        value2.ScheduleId
	Name      value2.ScheduleName
	EndAt     value2.ScheduleEndAt
	Period    value2.SchedulePeriod
	Timetable value2.ScheduleTimeTable
}
