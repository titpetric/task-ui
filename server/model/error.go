package model

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

// ErrorResponse renderer type for errors.
type ErrorResponse struct {
	// Error holds the error value
	Err error `json:"-"`

	// StatusCode holds the http response code.
	StatusCode int `json:"-"`

	// Status is the readable http message of StatusCode.
	Status string `json:"status"`

	// Error holds the encoded error message.
	Error ErrorBody `json:"error,omitempty"`
}

// ErrorBody is the encoding for an error message.
type ErrorBody struct {
	// Message contains the string representation of an error.
	Message string `json:"message"`
}

// NewErrorResponse produces a filled ErrorResponse object.
func NewErrorResponse(code int, err error) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: code,
		Status:     http.StatusText(code),
		Err:        err,
		Error: ErrorBody{
			Message: err.Error(),
		},
	}
}

// Error sets the error message and returns a new ErrorResponse.
func (e *ErrorResponse) WithError(err error) *ErrorResponse {
	return NewErrorResponse(e.StatusCode, err)
}

// Render is a hook to print status code in header.
func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)
	return nil
}

func InternalServerError(err error) *ErrorResponse {
	return NewErrorResponse(http.StatusInternalServerError, err)
}

func NotFoundError(err error) *ErrorResponse {
	return NewErrorResponse(http.StatusNotFound, err)
}

var ErrNotFound = errors.New("record not found")
