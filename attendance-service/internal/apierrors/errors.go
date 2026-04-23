package apierrors

import "net/http"

// HTTP is an error with an explicit HTTP status for handlers.
type HTTP struct {
	Status int
	Msg    string
}

func (e *HTTP) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return http.StatusText(e.Status)
}

func Forbidden(msg string) error {
	return &HTTP{Status: http.StatusForbidden, Msg: msg}
}

func Conflict(msg string) error {
	return &HTTP{Status: http.StatusConflict, Msg: msg}
}

func BadRequest(msg string) error {
	return &HTTP{Status: http.StatusBadRequest, Msg: msg}
}

func NotFound(msg string) error {
	return &HTTP{Status: http.StatusNotFound, Msg: msg}
}

func ServiceUnavailable(msg string) error {
	return &HTTP{Status: http.StatusServiceUnavailable, Msg: msg}
}
