package types

import (
	"context"
	"io"
)

type Level int

const (
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel Level = iota - 1

	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel

	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel

	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel

	// PanicLevel level. Logs and then calls panic with the message passed to Debug, Info, ...
	PanicLevel
)

type Logger interface {
	// SetLevel resets enabled log level
	SetLevel(lvl Level)

	// SetWriter resets log writer
	SetWriter(w AsyncWriter)

	// AddCallerSkip increases the number of callers skipped by caller annotation
	// (as enabled by the AddCaller option). When building wrappers around the
	// Logger and SugaredLogger, supplying this Option prevents zap from always
	// reporting the wrapper code as the caller.
	AddCallerSkip(depth int) Logger

	// Stop flushes any buffered log entries, then stops writer.
	// This method should be called when program stopping.
	// For example,
	//   func main() {
	//     defer log.Stop()
	//
	//     log.Init(lvl, path)
	//   }
	Stop()

	// With adds a variadic number of fields to the logging context. It accepts
	// loosely-typed key-value pairs. When processing pairs, the first element
	// of the pair is used as the field key and the second as the field value.
	// Note that the keys in key-value pairs should be strings.
	// For example,
	//   log.With(
	//     "hello", "world",
	//     "failure", errors.New("oh no"),
	//     "count", 42,
	//     "user", User{Name: "alice"},
	//     "err", errors.New("error")
	//  )
	With(kvs ...interface{}) Logger

	// Debug uses fmt.Sprint to construct and log a message.
	Debug(args ...interface{})
	// Info uses fmt.Sprint to construct and log a message.
	Info(args ...interface{})
	// Warn uses fmt.Sprint to construct and log a message.
	Warn(args ...interface{})
	// Error uses fmt.Sprint to construct and log a message.
	Error(args ...interface{})
	// Panic uses fmt.Sprint to construct and log a message, then panics.
	Panic(args ...interface{})

	// Debugf uses fmt.Sprintf to log a templated message.
	Debugf(fmt string, args ...interface{})
	// Infof uses fmt.Sprintf to log a templated message.
	Infof(fmt string, args ...interface{})
	// Warnf uses fmt.Sprintf to log a templated message.
	Warnf(fmt string, args ...interface{})
	// Errorf uses fmt.Sprintf to log a templated message.
	Errorf(fmt string, args ...interface{})
	// Panicf uses fmt.Sprintf to log a templated message, then panics.
	Panicf(fmt string, args ...interface{})

	// Debugw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	//
	// When debug-level logging is disabled, this is much faster than
	//  s.With(keysAndValues).Debug(msg)
	Debugw(msg string, kvs ...interface{})
	// Infow logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Infow(msg string, kvs ...interface{})
	// Warnw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Warnw(msg string, kvs ...interface{})
	// Errorw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Errorw(msg string, kvs ...interface{})
	// Panicw logs a message with some additional context, then panics. The
	// variadic key-value pairs are treated as they are in With.
	Panicw(msg string, kvs ...interface{})

	// CtxDebug uses fmt.Sprint to construct and log a message. The metainfo in Context
	// will be appended to message as fields.
	CtxDebug(ctx context.Context, args ...interface{})
	// CtxInfo uses fmt.Sprint to construct and log a message. The metainfo in Context
	// will be appended to message as fields.
	CtxInfo(ctx context.Context, args ...interface{})
	// CtxWarn uses fmt.Sprint to construct and log a message. The metainfo in Context
	// will be appended to message as fields.
	CtxWarn(ctx context.Context, args ...interface{})
	// CtxError uses fmt.Sprint to construct and log a message. The metainfo in Context
	// will be appended to message as fields.
	CtxError(ctx context.Context, args ...interface{})
	// CtxPanic uses fmt.Sprint to construct and log a message, then panics. The metainfo
	// in Context will be appended to message as fields.
	CtxPanic(ctx context.Context, args ...interface{})

	// CtxDebugf uses fmt.Sprintf to log a templated message. The metainfo in Context
	// will be appended to message as fields.
	CtxDebugf(ctx context.Context, fmt string, args ...interface{})
	// CtxInfof uses fmt.Sprintf to log a templated message. The metainfo in Context
	// will be appended to message as fields.
	CtxInfof(ctx context.Context, fmt string, args ...interface{})
	// CtxWarnf uses fmt.Sprintf to log a templated message. The metainfo in Context
	// will be appended to message as fields.
	CtxWarnf(ctx context.Context, fmt string, args ...interface{})
	// CtxErrorf uses fmt.Sprintf to log a templated message. The metainfo in Context
	// will be appended to message as fields.
	CtxErrorf(ctx context.Context, fmt string, args ...interface{})
	// CtxPanicf uses fmt.Sprintf to log a templated message, then panics. The metainfo
	// in Context will be appended to message as fields.
	CtxPanicf(ctx context.Context, fmt string, args ...interface{})

	// CtxDebugw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With. The metainfo in Context will be appended to
	// message as fields.
	//
	// When debug-level logging is disabled, this is much faster than
	//  s.With(keysAndValues).CtxDebug(msg)
	CtxDebugw(ctx context.Context, msg string, kvs ...interface{})
	// CtxInfow logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With. The metainfo in Context will be appended to
	// message as fields.
	CtxInfow(ctx context.Context, msg string, kvs ...interface{})
	// CtxWarnw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With. The metainfo in Context will be appended to
	// message as fields.
	CtxWarnw(ctx context.Context, msg string, kvs ...interface{})
	// CtxErrorw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With. The metainfo in Context will be appended to
	// message as fields.
	CtxErrorw(ctx context.Context, msg string, kvs ...interface{})
	// CtxPanicw logs a message with some additional context, then panics. The variadic key-value
	// pairs are treated as they are in With. The metainfo in Context will be appended to
	// message as fields.
	CtxPanicw(ctx context.Context, msg string, kvs ...interface{})
}

type AsyncWriter interface {
	io.Writer

	Sync() error

	// Stop flush all log entries
	Stop() error
}
