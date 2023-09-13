package gfsptask

import (
	"encoding/hex"
	"fmt"
	"strconv"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
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
	KeyPrefixGfSpGCBucketMigrationTask      = "GCBucketMigration"
)

func GfSpCreateBucketApprovalTaskKey(bucket string, account string, fingerprint []byte) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpCreateBucketApprovalTask + CombineKey("bucket:"+bucket, "account:"+account,
		"fingerprint:"+hex.EncodeToString(fingerprint)))
}

func GfSpCreateObjectApprovalTaskKey(bucket, object string, account string, fingerprint []byte) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpCreateObjectApprovalTask + CombineKey("bucket:"+bucket, "object:"+object,
		"account:"+account, "fingerprint:"+hex.EncodeToString(fingerprint)))
}

func GfSpMigrateBucketApprovalTaskKey(bucket string, migrateBucketHash string) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpMigrateBucketApprovalTask + CombineKey("bucket:"+bucket, "hash:"+migrateBucketHash))
}
func GfSpReplicatePieceApprovalTaskKey(bucket, object, id string) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpReplicatePieceApprovalTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpDownloadObjectTaskKey(bucket, object, id string, low, high int64) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpDownloadObjectTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id,
		"low:"+fmt.Sprint(low), "high:"+fmt.Sprint(high)))
}

func GfSpDownloadPieceTaskKey(bucket, object, pieceKey string, pieceOffset, pieceLength uint64) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpDownloadPieceTask + CombineKey("bucket:"+bucket, "object:"+object, "piece:"+pieceKey,
		"offset:"+fmt.Sprint(pieceOffset), "length:"+fmt.Sprint(pieceLength)))
}

func GfSpChallengePieceTaskKey(bucket, object, id string, sIdx uint32, rIdx int32, user string) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpChallengeTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id,
		"sIdx:"+fmt.Sprint(sIdx), "rIdx:"+fmt.Sprint(rIdx), user))
}

func GfSpUploadObjectTaskKey(bucket, object, id string) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpUploadObjectTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpResumableUploadObjectTaskKey(bucket, object, id string, offset uint64) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpResumableUploadObjectTask + CombineKey(bucket, object, id, strconv.FormatUint(offset, 10)))
}

func GfSpReplicatePieceTaskKey(bucket, object, id string) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpReplicatePieceTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpRecoverPieceTaskKey(bucket, object, id string, pIdx uint32, replicateIdx int32, time int64) coretask.TKey {
	if replicateIdx >= 0 {
		return coretask.TKey(KeyPrefixGfSpRecoverPieceTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id,
			"segIdx:"+fmt.Sprint(pIdx), "ecIdx:"+fmt.Sprint(replicateIdx), "time"+fmt.Sprint(time)))
	}
	return coretask.TKey(KeyPrefixGfSpRecoverPieceTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id,
		"segIdx:"+fmt.Sprint(pIdx), "time"+fmt.Sprint(time)))
}

func GfSpSealObjectTaskKey(bucket, object, id string) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpSealObjectTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id))
}

func GfSpReceivePieceTaskKey(bucket, object, id string, rIdx uint32, pIdx int32) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpReceivePieceTask + CombineKey("bucket:"+bucket, "object:"+object, "id:"+id,
		"segmentIdx:"+fmt.Sprint(rIdx), "redundancyIdx:"+fmt.Sprint(pIdx)))
}

func GfSpGCObjectTaskKey(start, end uint64, time int64) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpGCObjectTask + CombineKey("start"+fmt.Sprint(start), "end"+fmt.Sprint(end),
		"time"+fmt.Sprint(time)))
}

func GfSpGCZombiePieceTaskKey(time int64) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpGCZombiePieceTask + CombineKey("time"+fmt.Sprint(time)))
}

func GfSpGfSpGCMetaTaskKey(time int64) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpGfSpGCMetaTask + CombineKey("time"+fmt.Sprint(time)))
}

func GfSpMigrateGVGTaskKey(oldGvgID uint32, bucketID uint64, redundancyIndex int32) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpMigrateGVGTask + CombineKey("oldGvgID"+fmt.Sprint(oldGvgID),
		"bucketID"+fmt.Sprint(bucketID), "redundancyIndex"+fmt.Sprint(redundancyIndex)))
}

func GfSpMigratePieceTaskKey(object, id string, redundancyIdx uint32, ecIdx int32) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpMigratePieceTask + CombineKey("object:"+object, "id:", id, "segIdx:",
		fmt.Sprint(redundancyIdx), "redundancyIndex:", fmt.Sprint(ecIdx)))
}

func GfSpGCBucketMigrationTaskKey(bucketID uint64) coretask.TKey {
	return coretask.TKey(KeyPrefixGfSpGCBucketMigrationTask + CombineKey("bucketID"+fmt.Sprint(bucketID)))
}

func CombineKey(field ...string) string {
	key := ""
	for _, f := range field {
		key = key + Delimiter + f
	}
	return key
}
