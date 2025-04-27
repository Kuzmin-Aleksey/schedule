package util

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func InsertFunc[S ~[]E, E any](s S, v E, f func(E) bool) S {
	if len(s) == 0 {
		return []E{v}
	}
	for i := 0; i < len(s); i++ {
		if f(s[i]) {
			return append(s[:i], append([]E{v}, s[i:]...)...)
		}
	}

	return append(s, v)
}

func ParseInt(s string) (int, bool) {
	var durInt int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		durInt = durInt*10 + int(r-'0')
	}
	return durInt, true
}

func ParseTimezone(s string) (*time.Location, error) {
	tz, err := time.Parse("-07:00", s)
	if err != nil {
		return nil, err
	}
	return tz.Location(), nil
}

type JsonDuration time.Duration

func (d *JsonDuration) MarshalJSON() ([]byte, error) {
	return []byte("\"" + time.Duration(*d).String() + "\""), nil
}

func (d *JsonDuration) UnmarshalJSON(data []byte) error {
	v := string(data)
	v = strings.Trim(v, "\"")
	// try parse int
	if dur, ok := ParseInt(v); ok {
		*d = JsonDuration(time.Duration(dur) * time.Second)
		return nil
	}

	dur, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*d = JsonDuration(dur)
	return nil
}

func Ptr[T any](v T) *T {
	return &v
}

func ConvertStructToMap(s any) map[string]any {
	m := make(map[string]any)

	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)

	for i := range t.NumField() {
		f := t.Field(i)
		val := v.Field(i)

		if !val.CanInterface() {
			continue
		}
		
		if val.Type().Kind() == reflect.Struct {
			m2 := ConvertStructToMap(val.Interface())
			if len(m2) > 0 {
				m[f.Name] = m2
			} else {
				m[f.Name] = fmt.Sprint(val.Interface())
			}
			continue
		}

		m[f.Name] = val.Interface()

	}

	return m
}
