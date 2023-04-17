package rcmgr

import (
	"errors"
)

// ErrResourceLimitExceeded is returned when attempting to perform an operation that would
// exceed system resource limits.
var ErrResourceLimitExceeded = errors.New("resource limit exceeded")

// ErrResourceScopeClosed is returned when attempting to reserve resources in a closed resource
// scope.
var ErrResourceScopeClosed = errors.New("resource scope closed")

type ErrMemoryLimitExceeded struct {
	current, attempted, limit int64
	priority                  uint8
	err                       error
}

func (e *ErrMemoryLimitExceeded) Error() string { return e.err.Error() }
func (e *ErrMemoryLimitExceeded) Unwrap() error { return e.err }

// edge may be empty if this is not an edge error
func logValuesMemoryLimit(scope, edge string, stat ScopeStat, err error) []interface{} {
	logValues := make([]interface{}, 0, 2*8)
	logValues = append(logValues, "scope", scope)
	if edge != "" {
		logValues = append(logValues, "edge", edge)
	}
	var e *ErrMemoryLimitExceeded
	if errors.As(err, &e) {
		logValues = append(logValues,
			"current", e.current,
			"attempted", e.attempted,
			"priority", e.priority,
			"limit", e.limit,
		)
	}
	return append(logValues, "stat", stat, "error", err)
}

type ErrConnLimitExceeded struct {
	current, attempted, limit int
	err                       error
}

func (e *ErrConnLimitExceeded) Error() string { return e.err.Error() }
func (e *ErrConnLimitExceeded) Unwrap() error { return e.err }

// edge may be empty if this is not an edge error
func logValuesConnLimit(scope, edge string, dir Direction, stat ScopeStat, err error) []interface{} {
	logValues := make([]interface{}, 0, 2*9)
	logValues = append(logValues, "scope", scope)
	if edge != "" {
		logValues = append(logValues, "edge", edge)
	}
	logValues = append(logValues, "direction", dir)
	var e *ErrConnLimitExceeded
	if errors.As(err, &e) {
		logValues = append(logValues,
			"current", e.current,
			"attempted", e.attempted,
			"limit", e.limit,
		)
	}
	return append(logValues, "stat", stat, "error", err)
}

type ErrTaskLimitExceeded struct {
	priority                  ReserveTaskPriority
	current, attempted, limit int
	err                       error
}

func (e *ErrTaskLimitExceeded) Error() string { return e.err.Error() }
func (e *ErrTaskLimitExceeded) Unwrap() error { return e.err }

// edge may be empty if this is not an edge error
func logValuesTaskLimit(scope, edge string, stat ScopeStat, err error) []interface{} {
	logValues := make([]interface{}, 0, 2*8)
	logValues = append(logValues, "scope", scope)
	if edge != "" {
		logValues = append(logValues, "edge", edge)
	}
	var e *ErrTaskLimitExceeded
	if errors.As(err, &e) {
		logValues = append(logValues,
			"priority", e.priority,
			"current", e.current,
			"attempted", e.attempted,
			"limit", e.limit,
		)
	}
	return append(logValues, "stat", stat, "error", err)
}
