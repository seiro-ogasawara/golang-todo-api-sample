package utility

import "net/http"

type HTTPError struct {
	errCode int
	message string
	cause   error
}

func (e *HTTPError) Error() string {
	if e.message != "" {
		return e.message
	}
	return e.cause.Error()
}

func (e *HTTPError) ErrCode() int {
	return e.errCode
}

func (e *HTTPError) Unwrap() error {
	return e.cause
}

func NewHTTPError(errCode int, message string, cause error) *HTTPError {
	return &HTTPError{errCode, message, cause}
}

func BadRequest(message string, cause error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, message, cause)
}

func NotFound(message string, cause error) *HTTPError {
	return NewHTTPError(http.StatusNotFound, message, cause)
}
