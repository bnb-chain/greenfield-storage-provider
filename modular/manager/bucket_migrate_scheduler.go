package manager

import (
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
)

type GlobalVirtualGroupByBucketMigrateExecuteUnit struct {
	bucketID uint64
	GlobalVirtualGroupMigrateExecuteUnit
}

type BucketMigrateExecutePlan struct {
	virtualGroupManager            vgmgr.VirtualGroupManager
	PrimaryGVGByBucketMigrateUnits []*GlobalVirtualGroupByBucketMigrateExecuteUnit // bucket migrate，primary gvg
}

func (plan *BucketMigrateExecutePlan) loadFromDB() error {
	// subscribe progress
	// plan progress
	// task progress
	return nil
}
func (plan *BucketMigrateExecutePlan) storeToDB() error {
	// TODO:
	return nil
}

func (plan *BucketMigrateExecutePlan) updateProgress() error {
	// TODO: update memory and db.
	return nil
}

func (plan *BucketMigrateExecutePlan) startDestSPSchedule() {
	// TODO:
	// dispatch to task-dispatcher
}

func (b *BucketMigrateExecutePlan) Init() error {
	// TODO:
	return nil
}

func (b *BucketMigrateExecutePlan) Start() error {
	// TODO:
	return nil
}

// BucketMigrateScheduler subscribes bucket migrate events and produces a gvg migrate plan.
// TODO: support multiple buckets migrate
type BucketMigrateScheduler struct {
	manager                     *ManageModular
	spID                        uint32
	currentSubscribeBlockHeight int // load from db
	executePlan                 *BucketMigrateExecutePlan
}

func (b *BucketMigrateScheduler) Init() error {
	// TODO:
	return nil
}

func (b *BucketMigrateScheduler) Start() error {
	// TODO:
	go b.subscribeEvents()
	return nil
}

func (b *BucketMigrateScheduler) subscribeEvents() {
	subscribeBucketMigrateEventsTicker := time.NewTicker(time.Duration(b.manager.subscribeBucketMigrateEventInterval) * time.Second)
	for {
		select {
		case <-subscribeBucketMigrateEventsTicker.C:
			// spExitEvent, err = s.manager.baseApp.GfSpClient().ListBucketMigrateEvents(s.currentSubscribeBlockHeight, s.manager.baseApp.OperatorAddress())
			// TODO:
			// 1.subscribe metadata event
			// 1.make plan, start plan
			// 2.update plan
		}
	}
}
