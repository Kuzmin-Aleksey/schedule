package entity

import (
	"log/slog"
	"time"
)

const (
	MaxMedicineNameLen = 255
	MinSchedulePeriod  = time.Hour
	MaxSchedulePeriod  = time.Hour * 24
)

type Schedule struct {
	Id     int           `json:"id" db:"id"`
	UserId int64         `json:"user_id" db:"user_id"` // med police 16 digits, always int64
	Name   string        `json:"name" db:"name"`
	EndAt  *time.Time    `json:"end_at" db:"end_at"`
	Period time.Duration `json:"period" db:"period"`
}

func (s Schedule) LogValue() slog.Value {
	return slog.AnyValue(struct {
		Id     int           `json:"id"`
		Name   string        `json:"name"`
		EndAt  *time.Time    `json:"end_at"`
		Period time.Duration `json:"period"`
	}{
		Id:     s.Id,
		Name:   s.Name,
		EndAt:  s.EndAt,
		Period: s.Period,
	})
}
