package server

type CustomError interface {
	Error() string
}

type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}
