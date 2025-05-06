package value

type ScheduleName string

func (s ScheduleName) String() string {
	return string(s)
}
