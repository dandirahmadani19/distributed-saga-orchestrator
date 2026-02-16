package errors

type Code string

const (
	NotFound     Code = "NOT_FOUND"
	Conflict     Code = "CONFLICT"
	Invalid      Code = "INVALID"
	Forbidden    Code = "FORBIDDEN"
	Unauthorized Code = "UNAUTHORIZED"
	Internal     Code = "INTERNAL"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func E(code Code, msg string, err error) error {
	return &Error{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}
