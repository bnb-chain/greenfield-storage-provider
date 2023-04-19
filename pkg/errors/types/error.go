package types

import (
	"errors"
	"fmt"
	"io"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
)

const (
	success = "success"
	unknown = "unknown error"
)

// Error returns an error with code and reason
func Error(code int, reason string) error {
	return &ServiceError{Code: uint32(code), Reason: reason}
}

// Errorf returns an error with format string
func Errorf(code int, format string, args ...any) error {
	return Error(code, fmt.Sprintf(format, args...))
}

// Error implements Error interface
func (se *ServiceError) Error() string {
	return fmt.Sprintf("code: %d, caused by: %s", se.GetCode(), se.GetReason())
}

// ErrCode returns error code
func (se *ServiceError) ErrCode() int {
	return int(se.GetCode())
}

// ErrReason returns internal error
func (se *ServiceError) ErrReason() string {
	return se.GetReason()
}

// Format is used to format error which implements fmt.Formatter
func (se *ServiceError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if se.GetReason() != "" {
				reason := fmt.Sprintf("code: %d, reason: %s", se.GetCode(), se.GetReason())
				_, _ = io.WriteString(s, reason)
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, se.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", se.Error())
	default:
		_, _ = fmt.Fprintf(s, "%%!%c(errs.Error=%s)", verb, se.Error())
	}
}

// Code gets error code from an error
func Code(e error) int {
	if e == nil {
		return merrors.SuccessCode
	}
	err, ok := e.(*ServiceError)
	if !ok && !errors.As(e, &err) {
		return merrors.UnknownErrCode
	}
	if err == nil {
		return merrors.SuccessCode
	}
	return err.ErrCode()
}

// Reason returns the underlying cause of the error
func Reason(e error) string {
	if e == nil {
		return success
	}
	err, ok := e.(*ServiceError)
	if !ok && !errors.As(e, &err) {
		return unknown
	}
	if err == nil {
		return success
	}
	return err.ErrReason()
}
