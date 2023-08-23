package gfspapp

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	mockUser     = "mockUser"
	mockPassword = "mockPassword"
	mockAddress  = "mockAddress"
	mockDatabase = "mockDatabase"
)

var mockErr = errors.New("mock error")

type mockApprover struct {
	t *testing.T
}

func (ma mockApprover) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockApprover(ctrl)
	m.EXPECT().Name().Return(module.ApprovalModularName).AnyTimes()
	return m, nil
}

type mockApproverFailure struct {
	t *testing.T
}

func (ma mockApproverFailure) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	return nil, errors.New("mock error")
}

type mockAuthenticator struct {
	t *testing.T
}

func (ma mockAuthenticator) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockAuthenticator(ctrl)
	m.EXPECT().Name().Return(module.AuthenticationModularName).AnyTimes()
	return m, nil
}

type mockDownloader struct {
	t *testing.T
}

func (ma mockDownloader) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockDownloader(ctrl)
	m.EXPECT().Name().Return(module.DownloadModularName).AnyTimes()
	return m, nil
}

type mockExecutor struct {
	t *testing.T
}

func (ma mockExecutor) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockTaskExecutor(ctrl)
	m.EXPECT().Name().Return(module.ExecuteModularName).AnyTimes()
	return m, nil
}

type mockManager struct {
	t *testing.T
}

func (ma mockManager) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockManager(ctrl)
	m.EXPECT().Name().Return(module.ManageModularName).AnyTimes()
	return m, nil
}

type mockP2P struct {
	t *testing.T
}

func (ma mockP2P) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockP2P(ctrl)
	m.EXPECT().Name().Return(module.P2PModularName).AnyTimes()
	return m, nil
}

type mockReceiver struct {
	t *testing.T
}

func (ma mockReceiver) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockReceiver(ctrl)
	m.EXPECT().Name().Return(module.ReceiveModularName).AnyTimes()
	return m, nil
}

type mockSigner struct {
	t *testing.T
}

func (ma mockSigner) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockSigner(ctrl)
	m.EXPECT().Name().Return(module.SignModularName).AnyTimes()
	return m, nil
}

type mockUploader struct {
	t *testing.T
}

func (ma mockUploader) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockUploader(ctrl)
	m.EXPECT().Name().Return(module.UploadModularName).AnyTimes()
	return m, nil
}

type mockGater struct {
	t *testing.T
}

func (ma mockGater) new(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (module.Modular, error) {
	ctrl := gomock.NewController(ma.t)
	m := module.NewMockModular(ctrl)
	m.EXPECT().Name().Return(module.GateModularName).AnyTimes()
	return m, nil
}

func mockRegisterModular(t *testing.T) {
	approver := mockApprover{t: t}
	auth := mockAuthenticator{t: t}
	download := mockDownloader{t: t}
	execute := mockExecutor{t: t}
	gate := mockGater{t: t}
	manage := mockManager{t: t}
	p2p := mockP2P{t: t}
	receive := mockReceiver{t: t}
	sign := mockSigner{t: t}
	upload := mockUploader{t: t}
	RegisterModular(module.ApprovalModularName, module.ApprovalModularDescription, approver.new)
	RegisterModular(module.AuthenticationModularName, module.AuthenticationModularDescription, auth.new)
	RegisterModular(module.DownloadModularName, module.DownloadModularDescription, download.new)
	RegisterModular(module.ExecuteModularName, module.ExecuteModularDescription, execute.new)
	RegisterModular(module.GateModularName, module.GateModularDescription, gate.new)
	RegisterModular(module.ManageModularName, module.ManageModularDescription, manage.new)
	RegisterModular(module.P2PModularName, module.P2PModularDescription, p2p.new)
	RegisterModular(module.ReceiveModularName, module.ReceiveModularDescription, receive.new)
	RegisterModular(module.SignModularName, module.SignModularDescription, sign.new)
	RegisterModular(module.UploadModularName, module.UploadModularDescription, upload.new)
}
