package value

import (
	"database/sql/driver"
	"fmt"
	"schedule/internal/util"
	"time"
)

type ScheduleEndAt struct {
	*time.Time
}

func (t ScheduleEndAt) ToTime() time.Time {
	if t.Time != nil {
		return *t.Time
	}
	return time.Time{}
}

func (t ScheduleEndAt) String() string {
	if t.Time == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// NullableString for quick convert to rest model
func (t ScheduleEndAt) NullableString() *string {
	if t.Time != nil {
		return util.Ptr(t.String())
	}
	return nil
}

func (t ScheduleEndAt) MarshalJSON() ([]byte, error) {
	if t.Time == nil {
		return []byte("null"), nil
	}
	return []byte("\"" + t.String() + "\""), nil
}

func (t ScheduleEndAt) IsNil() bool {
	return t.Time == nil
}

func (t *ScheduleEndAt) Scan(v any) error {
	if v == nil {
		t.Time = nil
		return nil
	}
	timeV, ok := v.(time.Time)
	if ok {
		*t = NewScheduleEndAt(&timeV)
		return nil
	}

	return fmt.Errorf("'%v' (type %T) is not a time.Time ", v, v)
}

func (t ScheduleEndAt) Value() (driver.Value, error) {
	if t.Time == nil {
		return nil, nil
	}
	return *t.Time, nil
}

func NewScheduleEndAt(t *time.Time) ScheduleEndAt {
	return ScheduleEndAt{
		Time: t,
	}
}
