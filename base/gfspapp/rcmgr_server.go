package gfspapp

import (
	"context"
	"net/http"

	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsperrors"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfspserver"
)

var ErrFutureSupport = gfsperrors.Register(BaseCodeSpace, http.StatusNotFound, 995301, "future support")

var _ gfspserver.GfSpResourceServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpSetResourceLimit(context.Context, *gfspserver.GfSpSetResourceLimitRequest) (
	*gfspserver.GfSpSetResourceLimitResponse, error,
) {
	return &gfspserver.GfSpSetResourceLimitResponse{Err: ErrFutureSupport}, nil
}

func (g *GfSpBaseApp) GfSpQueryResourceLimit(context.Context, *gfspserver.GfSpQueryResourceLimitRequest) (
	*gfspserver.GfSpQueryResourceLimitResponse, error,
) {
	return &gfspserver.GfSpQueryResourceLimitResponse{Err: ErrFutureSupport}, nil
}
