package rcmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenServiceScope(t *testing.T) {
	limits, _ := NewLimitConfigFromToml("./testdata/limit.toml")
	rm, err := NewResourceManager(limits)
	assert.NoError(t, err)

	uploaderScope, err := rm.OpenService("uploader")
	assert.NoError(t, err)
	downloaderScope, err := rm.OpenService("downloader")
	assert.NoError(t, err)

	err = uploaderScope.ReserveMemory(100, ReservationPriorityAlways)
	assert.NoError(t, err)
	rm.ViewSystem(func(scope ResourceScope) error {
		system := scope.(*resourceScope)
		assert.Equal(t, int64(100), system.rc.memory)
		return nil
	})
	rm.ViewService("uploader", func(scope ResourceScope) error {
		uploader := scope.(*resourceScope)
		assert.Equal(t, int64(100), uploader.rc.memory)
		return nil
	})

	err = downloaderScope.ReserveMemory(100, ReservationPriorityAlways)
	assert.NoError(t, err)
	rm.ViewSystem(func(scope ResourceScope) error {
		system := scope.(*resourceScope)
		assert.Equal(t, int64(200), system.rc.memory)
		return nil
	})
	rm.ViewService("downloader", func(scope ResourceScope) error {
		uploader := scope.(*resourceScope)
		assert.Equal(t, int64(100), uploader.rc.memory)
		return nil
	})

	downRcmgr, err := downloaderScope.BeginSpan()
	assert.NoError(t, err)
	err = downRcmgr.ReserveMemory(100, ReservationPriorityAlways)
	assert.NoError(t, err)
	rm.ViewSystem(func(scope ResourceScope) error {
		system := scope.(*resourceScope)
		assert.Equal(t, int64(300), system.rc.memory)
		return nil
	})
	rm.ViewService("downloader", func(scope ResourceScope) error {
		uploader := scope.(*resourceScope)
		assert.Equal(t, int64(200), uploader.rc.memory)
		return nil
	})
	downRcmgr.Done()
	rm.ViewSystem(func(scope ResourceScope) error {
		system := scope.(*resourceScope)
		assert.Equal(t, int64(200), system.rc.memory)
		return nil
	})
	rm.ViewService("downloader", func(scope ResourceScope) error {
		uploader := scope.(*resourceScope)
		assert.Equal(t, int64(100), uploader.rc.memory)
		return nil
	})
}
