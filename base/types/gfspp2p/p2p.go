package gfspp2p

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&GfSpPing{}, "p2p/Ping", nil)
	cdc.RegisterConcrete(&GfSpPong{}, "p2p/Pong", nil)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterCodec(Amino)
}

// GetSignBytes returns the ping message bytes to sign over.
func (m *GfSpPing) GetSignBytes() []byte {
	fakeMsg := proto.Clone(m).(*GfSpPing)
	fakeMsg.Signature = []byte{}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

// GetSignBytes returns the pong message bytes to sign over.
func (m *GfSpPong) GetSignBytes() []byte {
	fakeMsg := proto.Clone(m).(*GfSpPong)
	fakeMsg.Signature = []byte{}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}
