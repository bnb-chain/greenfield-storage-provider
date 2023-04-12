package errors

import (
	"errors"
	"fmt"
	"io"
)

const Success = "success"

// APIError structure
type APIError struct {
	cause error
	code  int32
	msg   string
}

// New returns an error with code
func New(code int) error {
	return &APIError{code: int32(code)}
}

// Error returns an error with code and msg
func Error(code int, msg string) error {
	return &APIError{code: int32(code), msg: msg}
}

// Errorf wraps an error with format string
func Errorf(code int, format string, args ...any) error {
	return Error(code, fmt.Sprintf(format, args...))
}

// Wrap error with code and msg
func Wrap(cause error, code int, msg string) error {
	if cause == nil {
		return nil
	}
	return &APIError{cause: cause, code: int32(code), msg: msg}
}

// Wrapf wraps error with format string
func Wrapf(cause error, code int, format string, args ...any) error {
	return Wrap(cause, code, fmt.Sprintf(format, args...))
}

// Error implements error interface, returns error description
func (a *APIError) Error() string {
	if a == nil {
		return Success
	}
	if a.cause != nil {
		return fmt.Sprintf("code: %d, msg: %s, caused by: %s", a.code, a.msg, a.cause.Error())
	}
	return fmt.Sprintf("code: %d, msg: %s", a.code, a.msg)
}

// Code returns error code
func (a *APIError) Code() int {
	return int(a.code)
}

// Cause returns internal error
func (a *APIError) Cause() error {
	return a.cause
}

// Format is used to format error which implements fmt.Formatter
func (a *APIError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if a.msg != "" {
				msg := fmt.Sprintf("code: %d, msg: %s", a.code, a.msg)
				_, _ = io.WriteString(s, msg)
			}
			if a.cause != nil {
				_, _ = fmt.Fprintf(s, "\ncaused by %+v", a.Cause())
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, a.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", a.Error())
	default:
		_, _ = fmt.Fprintf(s, "%%!%c(errs.Error=%s)", verb, a.Error())
	}
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//	type causer interface {
//	       Cause() error
//	}
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(e error) error {
	type causer interface {
		Cause() error
	}

	for e != nil {
		cause, ok := e.(causer)
		if !ok {
			break
		}
		e = cause.Cause()
	}
	return e
}

// Code gets error code from an error
func Code(e error) int {
	if e == nil {
		return 0 // 0 represents success
	}
	err, ok := e.(*APIError)
	if !ok && errors.As(e, &err) {
		return 9999 // 9999 represents unknown error
	}
	if err == nil {
		return 0
	}
	return int(err.code)
}

// Message gets error message from an error
func Message(e error) string {
	if e == nil {
		return Success
	}
	err, ok := e.(*APIError)
	if !ok {
		return e.Error()
	}
	if err == (*APIError)(nil) {
		return Success
	}
	if err.Cause() != nil {
		return err.Error()
	}
	return err.msg
}

// Join error code and errors
func Join(code int, errs []error) error {
	var msg string
	for _, err := range errs {
		if err != nil {
			es := APIError{code: int32(Code(err)), cause: err}
			if len(msg) == 0 {
				msg += es.Error()
			} else {
				msg += ";" + es.Error()
			}
		}
	}
	if len(msg) > 0 {
		return Errorf(code, msg)
	}
	return nil
}
