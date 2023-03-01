package metadata

import (
	"context"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

// This package provides a set of methods to manage meta information in `context.Context`.
// Currently, it's based on `github.com/bytedance/gopkg/cloud/metainfo`.
// https://github.com/bytedance/gopkg/tree/main/cloud/metainfo

// TODO: baggage

func GetValue(ctx context.Context, k string) (v string, ok bool) {
	return metainfo.GetValue(ctx, k)
}

func WithValue(ctx context.Context, k, v string) context.Context {
	return metainfo.WithValue(ctx, k, v)
}
