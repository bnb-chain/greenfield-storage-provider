package https

import (
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
)

type SubError struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type Error struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Errors  []SubError `json:"errors,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code:%d message:%s errors:%+v", e.Code, e.Message, e.Errors)
}

func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func NewBadRequestError(format string, a ...any) *Error {
	return &Error{
		Code:    errors.ErrorCodeBadRequest,
		Message: fmt.Sprintf(format, a...),
	}
}

func NewInternalError(format string, a ...any) *Error {
	return &Error{
		Code:    errors.ErrorCodeInternalError,
		Message: fmt.Sprintf(format, a...),
	}
}

func NewNotFoundError(format string, a ...any) *Error {
	return &Error{
		Code:    errors.ErrorCodeNotFound,
		Message: fmt.Sprintf(format, a...),
	}
}
