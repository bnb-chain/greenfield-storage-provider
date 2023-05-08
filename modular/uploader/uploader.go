package uploader

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
)

const (
	UploadModularName        = "ploader"
	UploadModularDescription = "upload modular supports uploads object payload data to primary sp"
)

var _ module.Uploader = &UploadModular{}

type UploadModular struct {
	endpoint    string
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

func (u *UploadModular) Description() string {
	return UploadModularDescription
}

func (u *UploadModular) Endpoint() string {
	return u.endpoint
}

func (u *UploadModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
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

func (u *UploadModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
	return
}
