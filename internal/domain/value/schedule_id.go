package value

import (
	"fmt"
	"strconv"
)

type ScheduleId int

func ParseScheduleId(s string) (ScheduleId, error) {
	if s == "" {
		return 0, fmt.Errorf("empty schedule id")
	}

	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("strconv.Atoi(%s): %w", s, err)
	}

	return ScheduleId(id), nil
}
