package failure

type baseError struct {
	Msg string
}

func newBaseError(msg string) baseError {
	return baseError{
		Msg: msg,
	}
}

func (e baseError) Error() string {
	return e.Msg
}
