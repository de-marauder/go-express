package server

import (
	"log"
	"net"
)

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
		log.Panicln(err)
	}
}

func InternalServerError(conn net.Conn) {
	res := NewHTTPResponse(conn)
	buildErrorResponse(500, res)
	res.Send("")
}
func BadRequestError(conn net.Conn) {
	res := NewHTTPResponse(conn)
	buildErrorResponse(400, res)
	res.Send("")
}

func buildErrorResponse(code int, res *HTTPResponse) {
	res.StatusCode = code
	res.Body = HTTPStatusCodeMap[res.StatusCode] + "\n"
	res.Headers = make(map[string]string)
	res.Headers["Content-Type"] = "text/plain"
	parseBody(res)
}