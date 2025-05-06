package entity

import (
	"errors"
	"schedule/internal/domain/value"
	"time"
)

const (
	MaxMedicineNameLen = 255
	MinSchedulePeriod  = value.SchedulePeriod(time.Hour)
	MaxSchedulePeriod  = value.SchedulePeriod(time.Hour * 24)
)

type Schedule struct {
	Id     value.ScheduleId     `db:"id"`
	UserId value.UserId         `db:"user_id" json:"-"` // med police 16 digits, always int64
	Name   value.ScheduleName   `db:"name"`
	EndAt  value.ScheduleEndAt  `db:"end_at"`
	Period value.SchedulePeriod `db:"period"`
}

type ScheduleWithDuration struct {
	Id       value.ScheduleId
	UserId   value.UserId
	Name     value.ScheduleName
	Duration value.ScheduleDuration
	Period   value.SchedulePeriod
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
	Id         value.ScheduleId
	Name       value.ScheduleName
	EndAt      value.ScheduleEndAt
	Period     value.SchedulePeriod
	NextTaking value.ScheduleNextTaking
}

type ScheduleTimetable struct {
	Id        value.ScheduleId
	Name      value.ScheduleName
	EndAt     value.ScheduleEndAt
	Period    value.SchedulePeriod
	Timetable value.ScheduleTimeTable
}
