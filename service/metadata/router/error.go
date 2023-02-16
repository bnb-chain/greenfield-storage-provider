package router

import "fmt"

func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func NewBadRequestError(format string, a ...any) *Error {
	return &Error{
		Code:    ErrorCodeBadRequest,
		Message: fmt.Sprintf(format, a...),
	}
}

func NewInternalError(format string, a ...any) *Error {
	return &Error{
		Code:    ErrorCodeInternalError,
		Message: fmt.Sprintf(format, a...),
	}
}

func NewNotFoundError(format string, a ...any) *Error {
	return &Error{
		Code:    ErrorCodeNotFound,
		Message: fmt.Sprintf(format, a...),
	}
}
