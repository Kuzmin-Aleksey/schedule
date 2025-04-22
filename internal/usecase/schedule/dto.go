package schedule

import (
	"errors"
	"schedule/internal/entity"
	"time"
)

type CreateScheduleDTO struct {
	UserId   int64         `json:"user_id"`
	Name     string        `json:"name"`
	Duration uint          `json:"duration"` // days
	Period   time.Duration `json:"period" swaggertype:"string" example:"1h30m"`
}

func (dto *CreateScheduleDTO) Validate() error {
	switch {
	case dto.UserId == 0:
		return errors.New("user id is required")
	case dto.Name == "":
		return errors.New("name is required")
	case len(dto.Name) > entity.MaxMedicineNameLen:
		return errors.New("medicine name is too long")
	case dto.Period < entity.MinSchedulePeriod:
		return errors.New("period is too short")
	case dto.Period > entity.MaxSchedulePeriod:
		return errors.New("period is too long")
	}

	return nil
}

type CreateScheduleResponseDTO struct {
	Id int `json:"id"`
}

type ScheduleResponseDTO struct {
	Id        int           `json:"id"`
	Name      string        `json:"name"`
	EndAt     *time.Time    `json:"end_at,omitempty" example:"2025-04-21T22:00:00Z"`
	Period    time.Duration `json:"period" swaggertype:"string" example:"1h30m"`
	Timetable []time.Time   `json:"timetable" example:"2025-04-21T08:00:00Z"`
}

type NextTakingResponseDTO struct {
	Id         int           `json:"id"`
	Name       string        `json:"name"`
	EndAt      *time.Time    `json:"end_at,omitempty" example:"2025-04-21T22:00:00Z"`
	Period     time.Duration `json:"period" swaggertype:"string" example:"1h30m"`
	NextTaking time.Time     `json:"next_taking" example:"2025-04-21T08:00:00Z"`
}
