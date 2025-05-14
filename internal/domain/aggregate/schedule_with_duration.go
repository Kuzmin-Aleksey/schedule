package aggregate

import (
	"errors"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
)

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
	case len(t.Name) > entity.MaxMedicineNameLen:
		return errors.New("medicine name is too long")
	case t.Period < entity.MinSchedulePeriod:
		return errors.New("period is too short")
	case t.Period > entity.MaxSchedulePeriod:
		return errors.New("period is too long")
	}
	return nil
}
