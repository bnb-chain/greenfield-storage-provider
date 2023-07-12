package gfsptask

import (
	"fmt"
	"strconv"

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
	KeyPrefixGfSpReplicatePieceTask         = "Uploading"
	KeyPrefixGfSpSealObjectTask             = "Uploading"
	KeyPrefixGfSpResumableUploadObjectTask  = "ResumableUploading"
	KeyPrefixGfSpRecoverPieceTask           = "Recovering"
	KeyPrefixGfSpReceivePieceTask           = "ReceivePiece"
	KeyPrefixGfSpGCObjectTask               = "GCObject"
	KeyPrefixGfSpGCZombiePieceTask          = "GCZombiePiece"
	KeyPrefixGfSpGfSpGCMetaTask             = "GCMeta"
	KeyPrefixGfSpMigrateGVGTask             = "MigrateGVG"
	KeyPrefixGfSpMigratePieceTask           = "MigratePiece"
)

func GfSpCreateBucketApprovalTaskKey(bucket string, visibility int32) task.TKey {
	return task.TKey(KeyPrefixGfSpCreateBucketApprovalTask + CombineKey("bucket:"+bucket,
		"visibility:"+fmt.Sprint(visibility)))
}

func GfSpMigrateBucketApprovalTaskKey(bucket string, migrateBucketHash string) task.TKey {
	return task.TKey(KeyPrefixGfSpMigrateBucketApprovalTask + CombineKey("bucket:"+bucket, "hash:"+migrateBucketHash))
}

func GfSpCreateObjectApprovalTaskKey(bucket, object string, visibility int32) task.TKey {
	return task.TKey(KeyPrefixGfSpCreateObjectApprovalTask +
		CombineKey("bucket:"+bucket, "object:"+object,
			"visibility:"+fmt.Sprint(visibility)))
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

func GfSpResumableUploadObjectTaskKey(bucket, object, id string, offset uint64) task.TKey {
	return task.TKey(KeyPrefixGfSpResumableUploadObjectTask + CombineKey(bucket, object, id, strconv.FormatUint(offset, 10)))
}

func GfSpReplicatePieceTaskKey(bucket, object, id string) task.TKey {
	return task.TKey(KeyPrefixGfSpReplicatePieceTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpRecoverPieceTaskKey(bucket, object, id string, pIdx uint32, replicateIdx int32, time int64) task.TKey {
	if replicateIdx >= 0 {
		return task.TKey(KeyPrefixGfSpRecoverPieceTask +
			CombineKey("bucket:"+bucket, "object:"+object, "id:"+id, "segIdx:"+fmt.Sprint(pIdx), "ecIdx:"+fmt.Sprint(replicateIdx), "time"+fmt.Sprint(time)))
	}
	return task.TKey(KeyPrefixGfSpRecoverPieceTask +
		CombineKey("bucket:"+bucket, "object:"+object, "id:"+id, "segIdx:"+fmt.Sprint(pIdx), "time"+fmt.Sprint(time)))
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

func GfSpMigrateGVGTaskKey(oldGvgID uint32, bucketID uint64, redundancyIndex int32) task.TKey {
	return task.TKey(KeyPrefixGfSpMigrateGVGTask + CombineKey(
		"oldGvgID"+fmt.Sprint(oldGvgID), "bucketID"+fmt.Sprint(bucketID), "redundancyIndex"+fmt.Sprint(redundancyIndex)))
}

func GfSpMigratePieceTaskKey(object, id string, redundancyIdx uint32, ecIdx int32) task.TKey {
	return task.TKey(KeyPrefixGfSpMigratePieceTask + CombineKey("object:"+object, "id:", id, "segIdx:",
		fmt.Sprint(redundancyIdx), "redundancyIndex:", fmt.Sprint(ecIdx)))
}

func CombineKey(field ...string) string {
	key := ""
	for _, f := range field {
		key = key + Delimiter + f
	}
	return key
}
