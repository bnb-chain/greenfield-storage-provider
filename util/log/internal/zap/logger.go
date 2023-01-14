package zap

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/bnb-chain/inscription-storage-provider/util/log/internal/metadata"
	"github.com/bnb-chain/inscription-storage-provider/util/log/internal/types"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	baseCallDepth = 2

	LevelKey   = "l"
	TimeKey    = "t"
	CallerKey  = "caller"
	MessageKey = "msg"
)

var _ types.Logger = (*logger)(nil)

type logger struct {
	level  *levelEnabler
	writer *asyncWriterProxy

	base *zap.Logger
}

func NewLogger(lvl types.Level, w types.AsyncWriter) *logger {
	encoderConfig := zapcore.EncoderConfig{
		LevelKey:       LevelKey,
		TimeKey:        TimeKey,
		NameKey:        zapcore.OmitKey,
		CallerKey:      CallerKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     MessageKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	l := &logger{
		level:  newLevelEnabler(toZapLevel(lvl)),
		writer: newAsyncWriter(w),
		base:   nil,
	}
	core := zapcore.NewCore(newZapJSONEncoder(encoderConfig, false), l.writer, l.level)
	l.base = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(baseCallDepth),
		zap.AddStacktrace(zap.PanicLevel),
	)
	l.base.Sugar()
	return l
}

func toZapLevel(lvl types.Level) zapcore.Level {
	switch lvl {
	case types.DebugLevel:
		return zap.DebugLevel
	case types.InfoLevel:
		return zap.InfoLevel
	case types.WarnLevel:
		return zap.WarnLevel
	case types.ErrorLevel:
		return zap.ErrorLevel
	case types.PanicLevel:
		return zap.PanicLevel
	default:
		return zap.InfoLevel
	}
}

func (zl *logger) SetLevel(lvl types.Level) {
	zl.level.SetLevel(toZapLevel(lvl))
}

func (zl *logger) SetWriter(w types.AsyncWriter) {
	zl.writer.SetWriter(w)
}

func (zl *logger) AddCallerSkip(skip int) types.Logger {
	return &logger{
		level:  zl.level,
		writer: zl.writer,
		base:   zl.base.WithOptions(zap.AddCallerSkip(skip)),
	}
}

func (zl *logger) With(args ...interface{}) types.Logger {
	return &logger{
		level:  zl.level,
		writer: zl.writer,
		base:   zl.base.With(zl.sweetenFields(args, 0)...),
	}
}

func (zl *logger) Debug(args ...interface{}) {
	zl.log(zap.DebugLevel, nil, "", args, nil)
}

func (zl *logger) Info(args ...interface{}) {
	zl.log(zap.InfoLevel, nil, "", args, nil)
}

func (zl *logger) Warn(args ...interface{}) {
	zl.log(zap.WarnLevel, nil, "", args, nil)
}

func (zl *logger) Error(args ...interface{}) {
	zl.log(zap.ErrorLevel, nil, "", args, nil)
}

func (zl *logger) Panic(args ...interface{}) {
	zl.log(zap.PanicLevel, nil, "", args, nil)
}

func (zl *logger) Debugf(fmt string, args ...interface{}) {
	zl.log(zap.DebugLevel, nil, fmt, args, nil)
}

func (zl *logger) Infof(fmt string, args ...interface{}) {
	zl.log(zap.InfoLevel, nil, fmt, args, nil)
}

func (zl *logger) Warnf(fmt string, args ...interface{}) {
	zl.log(zap.WarnLevel, nil, fmt, args, nil)
}

func (zl *logger) Errorf(fmt string, args ...interface{}) {
	zl.log(zap.ErrorLevel, nil, fmt, args, nil)
}

func (zl *logger) Panicf(fmt string, args ...interface{}) {
	zl.log(zap.PanicLevel, nil, fmt, args, nil)
}

func (zl *logger) Debugw(msg string, kvs ...interface{}) {
	zl.log(zap.DebugLevel, nil, msg, nil, kvs)
}

func (zl *logger) Infow(msg string, kvs ...interface{}) {
	zl.log(zap.InfoLevel, nil, msg, nil, kvs)
}

func (zl *logger) Warnw(msg string, kvs ...interface{}) {
	zl.log(zap.WarnLevel, nil, msg, nil, kvs)
}

func (zl *logger) Errorw(msg string, kvs ...interface{}) {
	zl.log(zap.ErrorLevel, nil, msg, nil, kvs)
}

func (zl *logger) Panicw(msg string, kvs ...interface{}) {
	zl.log(zap.PanicLevel, nil, msg, nil, kvs)
}

func (zl *logger) CtxDebug(ctx context.Context, args ...interface{}) {
	zl.log(zap.DebugLevel, ctx, "", args, nil)
}

func (zl *logger) CtxInfo(ctx context.Context, args ...interface{}) {
	zl.log(zap.InfoLevel, ctx, "", args, nil)
}

func (zl *logger) CtxWarn(ctx context.Context, args ...interface{}) {
	zl.log(zap.WarnLevel, ctx, "", args, nil)
}

func (zl *logger) CtxError(ctx context.Context, args ...interface{}) {
	zl.log(zap.ErrorLevel, ctx, "", args, nil)
}

func (zl *logger) CtxPanic(ctx context.Context, args ...interface{}) {
	zl.log(zap.PanicLevel, ctx, "", args, nil)
}

func (zl *logger) CtxDebugf(ctx context.Context, fmt string, args ...interface{}) {
	zl.log(zap.DebugLevel, ctx, fmt, args, nil)
}

func (zl *logger) CtxInfof(ctx context.Context, fmt string, args ...interface{}) {
	zl.log(zap.InfoLevel, ctx, fmt, args, nil)
}

func (zl *logger) CtxWarnf(ctx context.Context, fmt string, args ...interface{}) {
	zl.log(zap.WarnLevel, ctx, fmt, args, nil)
}

func (zl *logger) CtxErrorf(ctx context.Context, fmt string, args ...interface{}) {
	zl.log(zap.ErrorLevel, ctx, fmt, args, nil)
}

func (zl *logger) CtxPanicf(ctx context.Context, fmt string, args ...interface{}) {
	zl.log(zap.PanicLevel, ctx, fmt, args, nil)
}

func (zl *logger) CtxDebugw(ctx context.Context, msg string, kvs ...interface{}) {
	zl.log(zap.DebugLevel, ctx, msg, nil, kvs)
}

func (zl *logger) CtxInfow(ctx context.Context, msg string, kvs ...interface{}) {
	zl.log(zap.InfoLevel, ctx, msg, nil, kvs)
}

func (zl *logger) CtxWarnw(ctx context.Context, msg string, kvs ...interface{}) {
	zl.log(zap.WarnLevel, ctx, msg, nil, kvs)
}

func (zl *logger) CtxErrorw(ctx context.Context, msg string, kvs ...interface{}) {
	zl.log(zap.ErrorLevel, ctx, msg, nil, kvs)
}

func (zl *logger) CtxPanicw(ctx context.Context, msg string, kvs ...interface{}) {
	zl.log(zap.PanicLevel, ctx, msg, nil, kvs)
}

func (zl *logger) Stop() {
	zl.writer.Stop()
}

func (zl *logger) log(lvl zapcore.Level, ctx context.Context, template string, fmtArgs []interface{}, kvs []interface{}) {
	if lvl < zap.DPanicLevel && !zl.base.Core().Enabled(lvl) {
		return
	}

	msg := zl.getMessage(template, fmtArgs)
	if ce := zl.base.Check(lvl, msg); ce != nil {
		fields := zl.getMetaInfo(ctx)
		fields = append(fields, zl.sweetenFields(kvs, 1)...)
		ce.Write(fields...)
	}
}

// getMetaInfo add these field
// - trace_id
func (zl *logger) getMetaInfo(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}

	fields := make([]zap.Field, 0)
	if traceID, ok := metadata.GetValue(ctx, "trace_id"); ok {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if objectID, ok := metadata.GetValue(ctx, "object_id"); ok {
		fields = append(fields, zap.String("object_id", objectID))
	}
	return fields
}

func (zl *logger) getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

const (
	_oddNumberErrMsg    = "Ignored key without a value."
	_nonStringKeyErrMsg = "Ignored key-value pairs with non-string keys."
)

func (zl *logger) sweetenFields(args []interface{}, skipCaller int) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	// Allocate enough space for the worst case; if users pass only structured
	// fields, we shouldn't penalize them with extra allocations.
	fields := make([]zap.Field, 0, len(args))
	var invalid invalidPairs

	usedFields := map[string]int{
		LevelKey:   1,
		TimeKey:    1,
		CallerKey:  1,
		MessageKey: 1,
	}

	for i := 0; i < len(args); {
		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			zl.base.WithOptions(zap.AddCallerSkip(skipCaller)).Error(_oddNumberErrMsg, zap.Any("ignored", args[i]))
			break
		}

		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			// Subsequent errors are likely, so allocate once up front.
			if cap(invalid) == 0 {
				invalid = make(invalidPairs, 0, len(args)/2)
			}
			invalid = append(invalid, invalidPair{i, key, val})
		} else {
			usedFields[keyStr] += 1
			if usedFields[keyStr] > 1 {
				keyStr += strconv.Itoa(usedFields[keyStr])

			}
			fields = append(fields, zap.Any(keyStr, val))
		}
		i += 2
	}

	// If we encountered any invalid key-value pairs, log an error.
	if len(invalid) > 0 {
		zl.base.WithOptions(zap.AddCallerSkip(skipCaller)).Error(_nonStringKeyErrMsg, zap.Array("invalid", invalid))
	}
	return fields
}

type invalidPair struct {
	position   int
	key, value interface{}
}

func (p invalidPair) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("position", int64(p.position))
	zap.Any("key", p.key).AddTo(enc)
	zap.Any("value", p.value).AddTo(enc)
	return nil
}

type invalidPairs []invalidPair

func (ps invalidPairs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	var err error
	for i := range ps {
		err = multierr.Append(err, enc.AppendObject(ps[i]))
	}
	return err
}

type levelEnabler struct {
	rawLevel int32
}

func newLevelEnabler(lvl zapcore.Level) *levelEnabler {
	return &levelEnabler{rawLevel: int32(lvl)}
}

func (e *levelEnabler) Enabled(lvl zapcore.Level) bool {
	return lvl >= zapcore.Level(atomic.LoadInt32(&e.rawLevel))
}

func (e *levelEnabler) SetLevel(lvl zapcore.Level) {
	atomic.StoreInt32(&e.rawLevel, int32(lvl))
}

type asyncWriterProxy struct {
	raw unsafe.Pointer
}

func newAsyncWriter(wr types.AsyncWriter) *asyncWriterProxy {
	w := &asyncWriterProxy{}
	atomic.StorePointer(&w.raw, unsafe.Pointer(&wr))
	return w
}

func (w *asyncWriterProxy) SetWriter(wr types.AsyncWriter) {
	atomic.StorePointer(&w.raw, unsafe.Pointer(&wr))
}

func (w *asyncWriterProxy) Writer() types.AsyncWriter {
	return *(*types.AsyncWriter)(atomic.LoadPointer(&w.raw))
}

func (w *asyncWriterProxy) Write(p []byte) (int, error) {
	return w.Writer().Write(p)
}

func (w *asyncWriterProxy) Sync() error {
	return w.Writer().Sync()
}

func (w *asyncWriterProxy) Stop() {
	w.Writer().Stop()
}
