package stonenode

import "github.com/bnb-chain/inscription-storage-provider/store/piecestore/piece"

type StoneNodeConfig struct {
	StoneHubServiceAddress string
	SyncerServiceAddress   string
	PieceConfig            *piece.PieceStoreConfig
}
