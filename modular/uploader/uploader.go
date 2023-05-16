package uploader

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/model"
)

var (
	UploadModularName        = model.UploadModular
	UploadModularDescription = model.SpServiceDesc[model.UploadModular]
)

var _ module.Uploader = &UploadModular{}

type UploadModular struct {
	baseApp     *gfspapp.GfSpBaseApp
	scope       rcmgr.ResourceScope
	uploadQueue taskqueue.TQueue
}

func (u *UploadModular) Name() string {
	return UploadModularName
}

func (u *UploadModular) Start(ctx context.Context) error {
	scope, err := u.baseApp.ResourceManager().OpenService(u.Name())
	if err != nil {
		return err
	}
	u.scope = scope
	return nil
}

func (u *UploadModular) Stop(ctx context.Context) error {
	u.scope.Release()
	return nil
}

func (u *UploadModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error) {
	span, err := u.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (u *UploadModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
	return
}
