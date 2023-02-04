package job

import (
	"sync"

	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
)

// ObjectInfoContext maintains the object info, goroutine safe.
type ObjectInfoContext struct {
	object *ptypesv1pb.ObjectInfo
	jobDB  jobdb.JobDBV2
	metaDB metadb.MetaDB
	mu     sync.RWMutex
}

// NewObjectInfoContext return the instance of ObjectInfoContext.
func NewObjectInfoContext(object *ptypesv1pb.ObjectInfo, jobDB jobdb.JobDBV2, metaDB metadb.MetaDB) *ObjectInfoContext {
	return &ObjectInfoContext{
		object: object,
		jobDB:  jobDB,
		metaDB: metaDB,
	}
}

// GetObjectInfo return the object info.
func (ctx *ObjectInfoContext) GetObjectInfo() *ptypesv1pb.ObjectInfo {
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
func (ctx *ObjectInfoContext) GetObjectRedundancyType() ptypesv1pb.RedundancyType {
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
	return ctx.jobDB.GetPrimaryJobV2(ctx.object.GetObjectId())
}

// GetSecondaryJob load the secondary piece job from db and return.
func (ctx *ObjectInfoContext) getSecondaryJob() ([]*jobdb.PieceJob, error) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.jobDB.GetSecondaryJobV2(ctx.object.GetObjectId())
}

// SetPrimaryPieceJobDone set the primary piece jod completed and update DB.
func (ctx *ObjectInfoContext) SetPrimaryPieceJobDone(job *jobdb.PieceJob) error {
	return ctx.jobDB.SetPrimaryPieceJobDoneV2(ctx.object.GetObjectId(), job)
}

// SetSecondaryPieceJobDone set the secondary piece jod completed and update DB.
func (ctx *ObjectInfoContext) SetSecondaryPieceJobDone(job *jobdb.PieceJob) error {
	return ctx.jobDB.SetSecondaryPieceJobDoneV2(ctx.object.GetObjectId(), job)
}

// SetSetIntegrityHash set integrity hash info to meta db.
func (ctx *ObjectInfoContext) SetSetIntegrityHash(meta *metadb.IntegrityMeta) error {
	return ctx.metaDB.SetIntegrityMeta(meta)
}
