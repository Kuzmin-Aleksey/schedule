package value

import (
	"fmt"
	"log/slog"
	"strconv"
)

type UserId int64

func ParseUserId(s string) (UserId, error) {
	if s == "" {
		return 0, fmt.Errorf("empty user id")
	}

	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("strconv.ParseInt(%s, 10, 64): %w", s, err)
	}

	return UserId(id), nil
}

func (id UserId) LogValue() slog.Value {
	return slog.StringValue("hidden")
}
