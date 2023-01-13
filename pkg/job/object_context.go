package job

import (
	"sync"

	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
)

// ObjectInfoContext maintains the object info, goroutine safe.
type ObjectInfoContext struct {
	object *types.ObjectInfo
	jobDB  jobdb.JobDB
	metaDB metadb.MetaDB
	mu     sync.RWMutex
}

// NewObjectInfoContext return the instance of ObjectInfoContext.
func NewObjectInfoContext(object *types.ObjectInfo, jobDB jobdb.JobDB, metaDB metadb.MetaDB) *ObjectInfoContext {
	return &ObjectInfoContext{
		object: object,
		jobDB:  jobDB,
		metaDB: metaDB,
	}
}

// GetObjectInfo return the object info.
func (ctx *ObjectInfoContext) GetObjectInfo() *types.ObjectInfo {
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
	return ctx.object.GetSize()
}

// GetObjectRedundancyType return the object redundancy type.
func (ctx *ObjectInfoContext) GetObjectRedundancyType() types.RedundancyType {
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
func (ctx *ObjectInfoContext) getPrimaryPieceJob() ([]*jobdb.PieceJob, error) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.jobDB.GetPrimaryJob(ctx.object.TxHash)
}

// GetSecondaryJob load the secondary piece job from db and return.
func (ctx *ObjectInfoContext) getSecondaryJob() ([]*jobdb.PieceJob, error) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.jobDB.GetSecondaryJob(ctx.object.TxHash)
}

// SetPrimaryPieceJobDone set the primary piece jod completed and update DB.
func (ctx *ObjectInfoContext) SetPrimaryPieceJobDone(job *jobdb.PieceJob) error {
	return ctx.jobDB.SetPrimaryPieceJobDone(ctx.object.GetTxHash(), job)
}

// SetSecondaryPieceJobDone set the secondary piece jod completed and update DB.
func (ctx *ObjectInfoContext) SetSecondaryPieceJobDone(job *jobdb.PieceJob) error {
	return ctx.jobDB.SetSecondaryPieceJobDone(ctx.object.GetTxHash(), job)
}

// SetSetIntegrityHash set integrity hash info to meta db.
func (ctx *ObjectInfoContext) SetSetIntegrityHash(meta *metadb.IntegrityMeta) error {
	return ctx.metaDB.SetIntegrityMeta(meta)
}
