package store

import (
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
)

type MetaDB interface {
	SetIntegrityHash(bucketName, objectName string, pieceJob *service.PieceJob) error
}
