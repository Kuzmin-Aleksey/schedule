package entity

import (
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
	UserId value.UserId         `db:"user_id" json:"-"`
	Name   value.ScheduleName   `db:"name"`
	EndAt  value.ScheduleEndAt  `db:"end_at"`
	Period value.SchedulePeriod `db:"period"`
}
