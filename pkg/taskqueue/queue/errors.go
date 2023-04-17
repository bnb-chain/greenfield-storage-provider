package queue

import (
	"errors"
)

var (
	ErrQueueExceeded   = errors.New("task queue exceed")
	ErrTaskRepeated    = errors.New("task repeated")
	ErrUnsupportedTask = errors.New("task priority unsupported among queue")
)
