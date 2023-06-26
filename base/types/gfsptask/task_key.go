package gfsptask

import (
	"fmt"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/core/task"
)

const (
	Delimiter                               = "-"
	KeyPrefixGfSpCreateBucketApprovalTask   = "CreateBucketApproval"
	KeyPrefixGfSpMigrateBucketApprovalTask  = "MigrateBucketApproval"
	KeyPrefixGfSpCreateObjectApprovalTask   = "CreateObjectApproval"
	KeyPrefixGfSpReplicatePieceApprovalTask = "ReplicatePieceApproval"
	KeyPrefixGfSpDownloadObjectTask         = "DownloadObject"
	KeyPrefixGfSpDownloadPieceTask          = "DownloadPiece"
	KeyPrefixGfSpChallengeTask              = "ChallengePiece"
	KeyPrefixGfSpUploadObjectTask           = "Uploading"
	KeyPrefixGfSpReplicatePieceTask         = "Replicating"
	KeyPrefixGfSpSealObjectTask             = "Sealing"
	KeyPrefixGfSpReceivePieceTask           = "ReceivePiece"
)

var (
	KeyPrefixGfSpGCObjectTask      = strings.ToLower("GCObject")
	KeyPrefixGfSpGCZombiePieceTask = strings.ToLower("GCZombiePiece")
	KeyPrefixGfSpGfSpGCMetaTask    = strings.ToLower("GCMeta")
)

func GfSpCreateBucketApprovalTaskKey(bucket string, createBucketHash string) task.TKey {
	return task.TKey(KeyPrefixGfSpCreateBucketApprovalTask + CombineKey("bucket:"+bucket, "hash:"+createBucketHash))
}

func GfSpMigrateBucketApprovalTaskKey(bucket string, migrateBucketHash string) task.TKey {
	return task.TKey(KeyPrefixGfSpMigrateBucketApprovalTask + CombineKey("bucket:"+bucket, "hash:"+migrateBucketHash))
}

func GfSpCreateObjectApprovalTaskKey(bucket, object string, createObjectHash string) task.TKey {
	return task.TKey(KeyPrefixGfSpCreateObjectApprovalTask +
		CombineKey("bucket:"+bucket, "object:"+object, "hash:"+createObjectHash))
}

func GfSpReplicatePieceApprovalTaskKey(bucket, object, id string) task.TKey {
	return task.TKey(KeyPrefixGfSpReplicatePieceApprovalTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpDownloadObjectTaskKey(bucket, object, id string, low, high int64) task.TKey {
	return task.TKey(KeyPrefixGfSpDownloadObjectTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id,
			"low:"+fmt.Sprint(low), "high:"+fmt.Sprint(high)))
}

func GfSpDownloadPieceTaskKey(bucket, object, pieceKey string, pieceOffset, pieceLength uint64) task.TKey {
	return task.TKey(KeyPrefixGfSpDownloadPieceTask +
		CombineKey("bucket:"+bucket, "object:"+object, "piece:"+pieceKey,
			"offset:"+fmt.Sprint(pieceOffset), "length:"+fmt.Sprint(pieceLength)))
}

func GfSpChallengePieceTaskKey(bucket, object, id string, sIdx uint32, rIdx int32, user string) task.TKey {
	return task.TKey(KeyPrefixGfSpChallengeTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id,
			"sIdx:"+fmt.Sprint(sIdx), "rIdx:"+fmt.Sprint(rIdx), user))
}

func GfSpUploadObjectTaskKey(bucket, object, id string) task.TKey {
	return task.TKey(KeyPrefixGfSpUploadObjectTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpReplicatePieceTaskKey(bucket, object, id string) task.TKey {
	return task.TKey(KeyPrefixGfSpReplicatePieceTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpSealObjectTaskKey(bucket, object, id string) task.TKey {
	return task.TKey(KeyPrefixGfSpSealObjectTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpReceivePieceTaskKey(bucket, object, id string, rIdx uint32, pIdx int32) task.TKey {
	return task.TKey(KeyPrefixGfSpReceivePieceTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id,
			"rIdx:"+fmt.Sprint(rIdx), "pIdx:"+fmt.Sprint(pIdx)))
}

func GfSpGCObjectTaskKey(start, end uint64, time int64) task.TKey {
	return task.TKey(KeyPrefixGfSpGCObjectTask + CombineKey(
		"start"+fmt.Sprint(start), "end"+fmt.Sprint(end), "time"+fmt.Sprint(time)))
}

func GfSpGCZombiePieceTaskKey(time int64) task.TKey {
	return task.TKey(KeyPrefixGfSpGCZombiePieceTask + CombineKey("time"+fmt.Sprint(time)))
}

func GfSpGfSpGCMetaTaskKey(time int64) task.TKey {
	return task.TKey(KeyPrefixGfSpGfSpGCMetaTask + CombineKey("time"+fmt.Sprint(time)))
}

func CombineKey(field ...string) string {
	key := ""
	for _, f := range field {
		key = key + Delimiter + f
	}
	return key
}
