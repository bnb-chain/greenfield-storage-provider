package sr25519

import (
	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/jsontypes"
)

const (
	PrivKeyName = "tendermint/PrivKeySr25519"
	PubKeyName  = "tendermint/PubKeySr25519"
)

func init() {
	jsontypes.MustRegister(PubKey{})
	jsontypes.MustRegister(PrivKey{})
}
