package metadata

import (
	"context"
)

// context key
const (
	CtxKeyTraceID      = "trace_id"
	CtxKeyTraceHash    = "create_hash"
	CtxKeyTraceSampled = "trace_sampled"
)

// GetTraceID get trace-id from context
// return empty if not found
func GetTraceID(ctx context.Context) (s string) {
	s, _ = GetValue(ctx, CtxKeyTraceID)
	return
}

func GetTraceHash(ctx context.Context) (s string) {
	s, _ = GetValue(ctx, CtxKeyTraceHash)
	return
}

func SetTraceID(ctx context.Context, id string) context.Context {
	return WithValue(ctx, CtxKeyTraceID, id)
}

func GetTraceSampled(ctx context.Context) (sampled, ok bool) {
	v, ok := GetValue(ctx, CtxKeyTraceSampled)
	return v == "true", ok
}

func SetTraceSampledTrue(ctx context.Context) context.Context {
	return WithValue(ctx, CtxKeyTraceSampled, "true")
}

func SetTraceSampledFalse(ctx context.Context) context.Context {
	return WithValue(ctx, CtxKeyTraceSampled, "false")
}
