package value

import (
	"fmt"
	"time"
)

type SchedulePeriod time.Duration

func ParseSchedulePeriod(s string) (SchedulePeriod, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("time.ParseDuration(%s): %w", s, err)
	}
	return SchedulePeriod(d), nil
}

func (t SchedulePeriod) String() string {
	return time.Duration(t).String()
}
