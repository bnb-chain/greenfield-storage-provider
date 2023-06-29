package gfspserver

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (g *GfSpMigratePiece) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&GfSpMigratePiece{
		ObjectInfo:    g.GetObjectInfo(),
		StorageParams: g.GetStorageParams(),
		ReplicateIdx:  g.GetReplicateIdx(),
		EcIdx:         g.GetEcIdx(),
	}))
}

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&GfSpMigratePiece{}, "p2p/ReplicatePieceApprovalTask", nil)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterCodec(Amino)
}
