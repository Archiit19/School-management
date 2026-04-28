package apierrors

import "net/http"

type HTTP struct {
	Status int
	Msg    string
}

func (e *HTTP) Error() string {
	return e.Msg
}

func BadRequest(msg string) error {
	return &HTTP{Status: http.StatusBadRequest, Msg: msg}
}

func Forbidden(msg string) error {
	return &HTTP{Status: http.StatusForbidden, Msg: msg}
}

func NotFound(msg string) error {
	return &HTTP{Status: http.StatusNotFound, Msg: msg}
}

func Conflict(msg string) error {
	return &HTTP{Status: http.StatusConflict, Msg: msg}
}

func ServiceUnavailable(msg string) error {
	return &HTTP{Status: http.StatusServiceUnavailable, Msg: msg}
}
