package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&Ping{}, "p2p/Ping", nil)
	cdc.RegisterConcrete(&Pong{}, "p2p/Pong", nil)
	cdc.RegisterConcrete(&GetApprovalRequest{}, "p2p/GetApprovalRequest", nil)
	cdc.RegisterConcrete(&GetApprovalResponse{}, "p2p/GetApprovalResponse", nil)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterCodec(Amino)
}
