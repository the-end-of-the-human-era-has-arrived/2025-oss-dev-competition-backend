package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var (
	ErrNotFound = NewError(http.StatusNotFound, WithMessage("the request resource not found"))
)

type Error struct {
	statusCode int
	message    string
}

func NewError(statusCode int, opts ...option) Error {
	err := Error{
		statusCode: statusCode,
		message:    "error occured",
	}

	for _, o := range opts {
		o(&err)
	}

	return err
}

type option func(*Error)

func WithError(err error) option {
	return func(e *Error) {
		e.message = err.Error()
	}
}

func WithMessage(msg string) option {
	return func(e *Error) {
		e.message = msg
	}
}

func (e Error) Error() (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg = `fail to marshal error`
		}
	}()

	buf := bytes.NewBuffer(make([]byte, 0, 256))
	if err := json.NewEncoder(buf).Encode(&e); err != nil {
		panic(err)
	}
	return buf.String()
}

func (e Error) StatusCode() int { return e.statusCode }

func (e Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		StatusCode int    `json:"statusCode"`
		StatusText string `json:"statusText"`
		Msg        string `json:"message"`
	}{
		StatusCode: e.statusCode,
		StatusText: http.StatusText(e.statusCode),
		Msg:        e.message,
	})
}
