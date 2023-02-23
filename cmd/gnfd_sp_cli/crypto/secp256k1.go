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
	fmt.Println(fmt.Sprintf("operator_address, private_key: %s, puclic_key: %s",
		hex.EncodeToString(operator), hex.EncodeToString(operator.PubKey().Bytes())))
	fmt.Println(fmt.Sprintf("funding_address, private_key: %s, puclic_key: %s",
		hex.EncodeToString(funding), hex.EncodeToString(funding.PubKey().Bytes())))
	fmt.Println(fmt.Sprintf("approval_address, private_key: %s, puclic_key: %s",
		hex.EncodeToString(approval), hex.EncodeToString(approval.PubKey().Bytes())))
	fmt.Println(fmt.Sprintf("seal_address, private_key: %s, puclic_key: %s",
		hex.EncodeToString(seal), hex.EncodeToString(seal.PubKey().Bytes())))

	fmt.Println("set env:")
	if c.Bool("e") {
		fmt.Println(fmt.Sprintf("operator_puclic_key: %s", hex.EncodeToString(seal.PubKey().Bytes())))
		if err := os.Setenv(model.StorageProvider, hex.EncodeToString(operator.PubKey().Bytes())); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(fmt.Sprintf("operator_private_key: %s", hex.EncodeToString(operator)))
		if err := os.Setenv("SIGNER_OPERATOR_PRIV_KEY", hex.EncodeToString(operator)); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(fmt.Sprintf("funding_private_key: %s", hex.EncodeToString(funding)))
		if err := os.Setenv("SIGNER_FUNDING_PRIV_KEY", hex.EncodeToString(funding)); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(fmt.Sprintf("approval_private_key: %s", hex.EncodeToString(approval)))
		if err := os.Setenv("SIGNER_APPROVAL_PRIV_KEY", hex.EncodeToString(approval)); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(fmt.Sprintf("seal_private_key: %s", hex.EncodeToString(seal)))
		if err := os.Setenv("SIGNER_SEAL_PRIV_KEY", hex.EncodeToString(seal)); err != nil {
			fmt.Println(err)
			return
		}
	}
}
