package entity

import (
	"time"
)

const (
	MaxMedicineNameLen = 255
	MinSchedulePeriod  = time.Hour
	MaxSchedulePeriod  = time.Hour * 24
)

type Schedule struct {
	Id     int           `db:"id"`
	UserId int64         `db:"user_id" json:"-"` // med police 16 digits, always int64
	Name   string        `db:"name"`
	EndAt  *time.Time    `db:"end_at"`
	Period time.Duration `db:"period"`
}
