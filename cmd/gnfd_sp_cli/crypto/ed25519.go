package crypto

import (
	"encoding/hex"
	"fmt"

	"github.com/urfave/cli"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/crypto/ed25519"
)

var CreateP2PCommand = cli.Command{
	Name:   "p2p.key.create",
	Usage:  "create p2p key pair",
	Action: CreateP2PAddress,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "n, number",
			Value: 1,
			Usage: "the number of create key pairs",
		},
	},
}

func CreateP2PAddress(c *cli.Context) {
	number := c.Int("n")
	for i := 0; i < number; i++ {
		nodeKey := ed25519.GenPrivKey()
		fmt.Printf("%d, private_key: %s, puclic_key: %s\n", i,
			hex.EncodeToString(nodeKey), hex.EncodeToString(nodeKey.PubKey().Bytes()))
	}
}
