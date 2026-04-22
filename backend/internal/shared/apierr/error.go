package apierr

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeInvalidRequest         Code = "INVALID_REQUEST"
	CodeUnsupportedMessageType Code = "UNSUPPORTED_MESSAGE_TYPE"
	CodeInvalidMessagePayload  Code = "INVALID_MESSAGE_PAYLOAD"
	CodeInvalidState           Code = "INVALID_STATE"
	CodeNotFound               Code = "NOT_FOUND"
	CodeForbidden              Code = "FORBIDDEN"
	CodeDuplicateEvent         Code = "DUPLICATE_EVENT"
	CodeInternalError          Code = "INTERNAL_ERROR"
)

type Error struct {
	HTTPStatus int
	Code       Code
	Message    string
	cause      error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

func New(code Code, msg string) *Error {
	return &Error{HTTPStatus: defaultStatus(code), Code: code, Message: msg}
}

func Wrap(code Code, msg string, cause error) *Error {
	return &Error{HTTPStatus: defaultStatus(code), Code: code, Message: msg, cause: cause}
}

func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func defaultStatus(c Code) int {
	switch c {
	case CodeInvalidRequest, CodeInvalidMessagePayload:
		return http.StatusBadRequest
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeDuplicateEvent:
		return http.StatusConflict
	case CodeInvalidState, CodeUnsupportedMessageType:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
