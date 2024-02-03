package server

import "log"

type CustomError interface {
	Error() string
}

type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

// Log Error and exit server
func logError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}