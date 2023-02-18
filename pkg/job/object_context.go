package job

import (
	"sync"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
)

// ObjectInfoContext maintains the object info, goroutine safe.
type ObjectInfoContext struct {
	object *ptypes.ObjectInfo
	jobDB  spdb.JobDB
	metaDB spdb.MetaDB
	mu     sync.RWMutex
}

// NewObjectInfoContext return the instance of ObjectInfoContext.
func NewObjectInfoContext(object *ptypes.ObjectInfo, jobDB spdb.JobDB, metaDB spdb.MetaDB) *ObjectInfoContext {
	return &ObjectInfoContext{
		object: object,
		jobDB:  jobDB,
		metaDB: metaDB,
	}
}

// GetObjectInfo return the object info.
func (ctx *ObjectInfoContext) GetObjectInfo() *ptypes.ObjectInfo {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.SafeCopy()
}

// GetObjectID return the object resource id.
func (ctx *ObjectInfoContext) GetObjectID() uint64 {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.GetObjectId()
}

// GetObjectSize return the object size.
func (ctx *ObjectInfoContext) GetObjectSize() uint64 {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.GetSize_()
}

// GetObjectRedundancyType return the object redundancy type.
func (ctx *ObjectInfoContext) GetObjectRedundancyType() ptypes.RedundancyType {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.GetRedundancyType()
}

// TxHash return the CreateObjectTX hash.
func (ctx *ObjectInfoContext) TxHash() []byte {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.GetTxHash()
}

// GetPrimaryJob load the primary piece job from db and return.
func (ctx *ObjectInfoContext) getPrimaryPieceJob() ([]*spdb.PieceJob, error) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.jobDB.GetPrimaryJob(ctx.object.GetObjectId())
}

// GetSecondaryJob load the secondary piece job from db and return.
func (ctx *ObjectInfoContext) getSecondaryJob() ([]*spdb.PieceJob, error) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.jobDB.GetSecondaryJob(ctx.object.GetObjectId())
}

// SetPrimaryPieceJobDone set the primary piece jod completed and update DB.
func (ctx *ObjectInfoContext) SetPrimaryPieceJobDone(job *spdb.PieceJob) error {
	return ctx.jobDB.SetPrimaryPieceJobDone(ctx.object.GetObjectId(), job)
}

// SetSecondaryPieceJobDone set the secondary piece jod completed and update DB.
func (ctx *ObjectInfoContext) SetSecondaryPieceJobDone(job *spdb.PieceJob) error {
	return ctx.jobDB.SetSecondaryPieceJobDone(ctx.object.GetObjectId(), job)
}

// SetSetIntegrityHash set integrity hash info to meta db.
func (ctx *ObjectInfoContext) SetSetIntegrityHash(meta *spdb.IntegrityMeta) error {
	return ctx.metaDB.SetIntegrityMeta(meta)
}
