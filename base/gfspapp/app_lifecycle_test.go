package gfspapp

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/zkMeLabs/mechain-storage-provider/core/module"
	"github.com/zkMeLabs/mechain-storage-provider/core/prober"
)

func TestGfSpBaseApp_startServicesSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := module.NewMockModular(ctrl)
	m.EXPECT().Start(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return nil
	}).AnyTimes()
	m.EXPECT().Name().Return("mock").AnyTimes()
	m1 := prober.NewMockProber(ctrl)
	m1.EXPECT().Ready().Return().AnyTimes()

	g := &GfSpBaseApp{}
	g.SetProbe(m1)
	g.RegisterServices(m)
	g.startServices(context.TODO())
}

func TestGfSpBaseApp_startServicesFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := module.NewMockModular(ctrl)
	m.EXPECT().Start(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return errors.New("mock error")
	}).AnyTimes()
	m.EXPECT().Name().Return("mock").AnyTimes()

	m1 := prober.NewMockProber(ctrl)
	m1.EXPECT().Unready(gomock.Any()).Return().AnyTimes()
	ctx, cancel := context.WithCancel(context.TODO())
	g := &GfSpBaseApp{appCtx: ctx, appCancel: cancel}
	g.SetProbe(m1)
	g.RegisterServices(m)
	g.startServices(context.TODO())
}

func TestGfSpBaseApp_stopServicesSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := module.NewMockModular(ctrl)
	m.EXPECT().Stop(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return nil
	}).AnyTimes()
	m.EXPECT().Name().Return("mock").AnyTimes()
	g := &GfSpBaseApp{}
	ctx, cancel := context.WithCancel(context.TODO())
	g.RegisterServices(m)
	g.stopServices(ctx, cancel)
}

func TestGfSpBaseApp_stopServicesFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := module.NewMockModular(ctrl)
	m.EXPECT().Stop(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return errors.New("mock error")
	}).AnyTimes()
	m.EXPECT().Name().Return("mock").AnyTimes()
	g := &GfSpBaseApp{}
	ctx, cancel := context.WithCancel(context.TODO())
	g.RegisterServices(m)
	g.stopServices(ctx, cancel)
}

func TestGfSpBaseApp_Done(t *testing.T) {
	g := &GfSpBaseApp{appCtx: context.TODO()}
	g.Done()
}
