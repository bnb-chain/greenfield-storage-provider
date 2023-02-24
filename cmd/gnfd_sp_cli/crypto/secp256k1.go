package crypto

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/urfave/cli"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/p2p/libs/common/crypto/secp256k1"
)

var CreateSpAddressCommand = cli.Command{
	Name:   "sp.address.create",
	Usage:  "create sp private and public address",
	Action: CreateSpAddress,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "e, set_env",
			Usage: "set sp private and public address to env vars",
		},
	},
}

func CreateSpAddress(c *cli.Context) {
	operator := secp256k1.GenPrivKey()
	funding := secp256k1.GenPrivKey()
	approval := secp256k1.GenPrivKey()
	seal := secp256k1.GenPrivKey()
	fmt.Printf("operator_address, private_key: %s, puclic_key: %s\n",
		hex.EncodeToString(operator), hex.EncodeToString(operator.PubKey().Bytes()))
	fmt.Printf("funding_address, private_key: %s, puclic_key: %s\n",
		hex.EncodeToString(funding), hex.EncodeToString(funding.PubKey().Bytes()))
	fmt.Printf("approval_address, private_key: %s, puclic_key: %s\n",
		hex.EncodeToString(approval), hex.EncodeToString(approval.PubKey().Bytes()))
	fmt.Printf("seal_address, private_key: %s, puclic_key: %s\n",
		hex.EncodeToString(seal), hex.EncodeToString(seal.PubKey().Bytes()))

	fmt.Print("set env:\n")
	if c.Bool("e") {
		fmt.Printf("operator_puclic_key: %s\n", hex.EncodeToString(seal.PubKey().Bytes()))
		if err := os.Setenv(model.StorageProvider, hex.EncodeToString(operator.PubKey().Bytes())); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("operator_private_key: %s\n", hex.EncodeToString(operator))
		if err := os.Setenv("SIGNER_OPERATOR_PRIV_KEY", hex.EncodeToString(operator)); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("funding_private_key: %s\n", hex.EncodeToString(funding))
		if err := os.Setenv("SIGNER_FUNDING_PRIV_KEY", hex.EncodeToString(funding)); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("approval_private_key: %s\n", hex.EncodeToString(approval))
		if err := os.Setenv("SIGNER_APPROVAL_PRIV_KEY", hex.EncodeToString(approval)); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("seal_private_key: %s\n", hex.EncodeToString(seal))
		if err := os.Setenv("SIGNER_SEAL_PRIV_KEY", hex.EncodeToString(seal)); err != nil {
			fmt.Println(err)
			return
		}
	}
}
