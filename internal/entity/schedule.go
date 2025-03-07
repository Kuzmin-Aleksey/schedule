package entity

import (
	"schedule/internal/util"
	"time"
)

const (
	MaxMedicineNameLen = 255
	MinSchedulePeriod  = util.JsonDuration(time.Hour)
	MaxSchedulePeriod  = util.JsonDuration(time.Hour * 24)
)

type Schedule struct {
	Id     int           `json:"id" db:"id"`
	UserId int64         `json:"user_id" db:"user_id"` // med police 16 digits, always int64
	Name   string        `json:"name" db:"name"`
	EndAt  *time.Time    `json:"end_at" db:"end_at"`
	Period time.Duration `json:"period" db:"period"`
}
