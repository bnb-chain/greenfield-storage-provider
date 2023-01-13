package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/bytedance/gopkg/cloud/metainfo"
	"go.uber.org/zap/zapcore"

	"github.com/bnb-chain/inscription-storage-provider/util/log/internal/types"
	"github.com/bnb-chain/inscription-storage-provider/util/log/internal/zap"
)

type (
	Logger = types.Logger
	Level  = types.Level
)

const (
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel = types.DebugLevel

	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel = types.InfoLevel

	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel = types.WarnLevel

	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel = types.ErrorLevel

	// PanicLevel level. Logs and then calls panic with the message passed to Debug, Info, ...
	PanicLevel = types.PanicLevel
)

// ParseLevel parses level from string.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "debug", "dbug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error", "eror":
		return ErrorLevel, nil
	case "panic":
		return PanicLevel, nil
	default:
		return InfoLevel, fmt.Errorf("invlaid level: %s", lvl)
	}
}

var logger Logger

func init() {
	logger = zap.NewLogger(InfoLevel, &writerWrapper{Writer: os.Stderr}).AddCallerSkip(1)
}

// NOTE: this func isn't thread safe
func SetLogger(newLogger Logger) {
	logger = newLogger
}

// AsyncWriter uses as log writer
type AsyncWriter types.AsyncWriter

type writerWrapper struct {
	io.Writer
}

func (w writerWrapper) Sync() error {
	return nil
}

func (w writerWrapper) Stop() error {
	return nil
}

// Init auto setting level and creating log directory
func Init(lvl Level, path string) {
	logger.SetLevel(lvl)

	if path != "" {
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				logger.With("err", err).Panicf("make log dir failed")
			}
		} else if err != nil {
			logger.With("err", err).Panicf("invalid dir stat")
		}

		logger.SetWriter(NewMultiWriteSyncer(NewAsyncFileWriter(path, 10*1024*1024),
			&zapcore.BufferedWriteSyncer{WS: os.Stdout, FlushInterval: time.Second}))
	}
}

// SetLevel resets enabled log level
func SetLevel(lvl Level) {
	logger.SetLevel(lvl)
}

// SetWriter resets log writer
func SetWriter(w types.AsyncWriter) {
	logger.SetWriter(w)
}

// Stop flushes any buffered log entries, then stops writer.
// This method should be called when program stopping.
// For example,
//
//	func main() {
//	  defer log.Stop()
//
//	  log.Init(lvl, path)
//	}
func Stop() {
	logger.Stop()
}

// With adds a variadic number of fields to the logging context. It accepts
// loosely-typed key-value pairs. When processing pairs, the first element
// of the pair is used as the field key and the second as the field value.
// Note that the keys in key-value pairs should be strings.
// For example,
//
//	 log.With(
//	   "hello", "world",
//	   "failure", errors.New("oh no"),
//	   "count", 42,
//	   "user", User{Name: "alice"},
//	   "err", errors.New("error")
//	)
func With(kvs ...interface{}) Logger {
	return logger.AddCallerSkip(-1).With(kvs...)
}

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Info uses fmt.Sprint to construct and log a message.
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(fmt string, args ...interface{}) {
	logger.Debugf(fmt, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(fmt string, args ...interface{}) {
	logger.Infof(fmt, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(fmt string, args ...interface{}) {
	logger.Warnf(fmt, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(fmt string, args ...interface{}) {
	logger.Errorf(fmt, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(fmt string, args ...interface{}) {
	logger.Panicf(fmt, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).Debug(msg)
func Debugw(msg string, kvs ...interface{}) {
	logger.Debugw(msg, kvs...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infow(msg string, kvs ...interface{}) {
	logger.Infow(msg, kvs...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Warnw(msg string, kvs ...interface{}) {
	logger.Warnw(msg, kvs...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Errorw(msg string, kvs ...interface{}) {
	logger.Errorw(msg, kvs...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicw(msg string, kvs ...interface{}) {
	logger.Panicw(msg, kvs...)
}

// CtxDebug uses fmt.Sprint to construct and log a message. The metaInfo in Context
// will be appended to message as fields.
func CtxDebug(ctx context.Context, args ...interface{}) {
	logger.CtxDebug(ctx, args...)
}

// CtxInfo uses fmt.Sprint to construct and log a message. The metaInfo in Context
// will be appended to message as fields.
func CtxInfo(ctx context.Context, args ...interface{}) {
	logger.CtxInfo(ctx, args...)
}

// CtxWarn uses fmt.Sprint to construct and log a message. The metaInfo in Context
// will be appended to message as fields.
func CtxWarn(ctx context.Context, args ...interface{}) {
	logger.CtxWarn(ctx, args...)
}

// CtxError uses fmt.Sprint to construct and log a message. The metaInfo in Context
// will be appended to message as fields.
func CtxError(ctx context.Context, args ...interface{}) {
	logger.CtxError(ctx, args...)
}

// CtxPanic uses fmt.Sprint to construct and log a message, then panics. The metaInfo
// in Context will be appended to message as fields.
func CtxPanic(ctx context.Context, args ...interface{}) {
	logger.CtxPanic(ctx, args...)
}

// CtxDebugf uses fmt.Sprintf to log a templated message. The metaInfo in Context
// will be appended to message as fields.
func CtxDebugf(ctx context.Context, fmt string, args ...interface{}) {
	logger.CtxDebugf(ctx, fmt, args...)
}

// CtxInfof uses fmt.Sprintf to log a templated message. The metaInfo in Context
// will be appended to message as fields.
func CtxInfof(ctx context.Context, fmt string, args ...interface{}) {
	logger.CtxInfof(ctx, fmt, args...)
}

// CtxWarnf uses fmt.Sprintf to log a templated message. The metaInfo in Context
// will be appended to message as fields.
func CtxWarnf(ctx context.Context, fmt string, args ...interface{}) {
	logger.CtxWarnf(ctx, fmt, args...)
}

// CtxErrorf uses fmt.Sprintf to log a templated message. The metaInfo in Context
// will be appended to message as fields.
func CtxErrorf(ctx context.Context, fmt string, args ...interface{}) {
	logger.CtxErrorf(ctx, fmt, args...)
}

// CtxPanicf uses fmt.Sprintf to log a templated message, then panics. The metaInfo
// in Context will be appended to message as fields.
func CtxPanicf(ctx context.Context, fmt string, args ...interface{}) {
	logger.CtxPanicf(ctx, fmt, args...)
}

// CtxDebugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. The metaInfo in Context will be appended to
// message as fields.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).CtxDebug(msg)
func CtxDebugw(ctx context.Context, msg string, kvs ...interface{}) {
	logger.CtxDebugw(ctx, msg, kvs...)
}

// CtxInfow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. The metaInfo in Context will be appended to
// message as fields.
func CtxInfow(ctx context.Context, msg string, kvs ...interface{}) {
	logger.CtxInfow(ctx, msg, kvs...)
}

// CtxWarnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. The metaInfo in Context will be appended to
// message as fields.
func CtxWarnw(ctx context.Context, msg string, kvs ...interface{}) {
	logger.CtxWarnw(ctx, msg, kvs...)
}

// CtxErrorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. The metaInfo in Context will be appended to
// message as fields.
func CtxErrorw(ctx context.Context, msg string, kvs ...interface{}) {
	logger.CtxErrorw(ctx, msg, kvs...)
}

// CtxPanicw logs a message with some additional context, then panics. The variadic key-value
// pairs are treated as they are in With. The metaInfo in Context will be appended to
// message as fields.
func CtxPanicw(ctx context.Context, msg string, kvs ...interface{}) {
	logger.CtxPanicw(ctx, msg, kvs...)
}

func Context(ctx context.Context, opts ...interface{}) context.Context {
	for _, req := range opts {
		if reflect.ValueOf(req).MethodByName("GetTraceId").IsValid() {
			valList := reflect.ValueOf(req).MethodByName("GetTraceId").Call([]reflect.Value{})
			if len(valList) > 0 && !valList[0].IsZero() {
				traceID := valList[0].String()
				ctx = metainfo.WithValue(ctx, "trace_id", traceID)
			}
		}
		if reflect.ValueOf(req).MethodByName("GetObjectId").IsValid() {
			valList := reflect.ValueOf(req).MethodByName("GetObjectId").Call([]reflect.Value{})
			if len(valList) > 0 && !valList[0].IsZero() {
				traceID := valList[0].String()
				ctx = metainfo.WithValue(ctx, "object_id", traceID)
			}
		}
	}
	return ctx
}

func WithValue(ctx context.Context, k, v string) context.Context {
	return metainfo.WithValue(ctx, k, v)
}
