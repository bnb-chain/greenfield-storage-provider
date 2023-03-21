package p2p

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli/v2"
)

var number = &cli.IntFlag{
	Name:  "n",
	Usage: "The number of key pairs",
	Value: 1,
}

var P2PCreateKeysCmd = &cli.Command{
	Action: p2pCreateKeysAction,
	Name:   "p2p.create.key",
	Usage:  "Create Secp256k1 key pairs for encrypting p2p protocol msg and identifying p2p node",
	Flags: []cli.Flag{
		number,
	},
	Category: "P2P COMMANDS",
	Description: `
The p2p.create.key creates 'n'' sets of Secp256k1 key pairs, each key pair contains a private key 
and a node id, the private key is used to encrypt p2p protocol msg, and the node id is use to public
to other p2p nodes for communication by p2p protocol.`,
}

func p2pCreateKeysAction(ctx *cli.Context) error {
	n := ctx.Int(number.Name)
	makeKeyPairs := func() (string, string, error) {
		privKey, _, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		if err != nil {
			return "", "", err
		}
		secp256k1PrivKey := privKey.(*crypto.Secp256k1PrivateKey)
		nodeID, err := peer.IDFromPublicKey(secp256k1PrivKey.GetPublic())
		if err != nil {
			return "", "", err
		}
		return secp256k1PrivKey.Key.String(), nodeID.String(), err
	}
	for i := 0; i < n; i++ {
		private, nodeId, err := makeKeyPairs()
		if err != nil {
			return err
		}
		fmt.Printf("%d private key: %s\n", i, private)
		fmt.Printf("%d node id key: %s\n", i, nodeId)
	}
	return nil
}
