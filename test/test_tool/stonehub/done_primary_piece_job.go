package stonehub

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/urfave/cli"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	cliCtx "github.com/bnb-chain/inscription-storage-provider/test/test_tool/context"
)

var DonePrimaryPieceJobCommand = cli.Command{
	Name:   "done_primary_piece_job",
	Usage:  "Complete primary piece job",
	Action: donePrimaryPieceJob,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "t,TxHash",
			Value: "",
			Usage: "Transaction hash of create object"},
		cli.StringFlag{
			Name:  "s,StorageProvider",
			Value: "",
			Usage: "StorageProvider id of primary"},
		cli.Uint64Flag{
			Name:  "i,PieceIdx",
			Value: 0,
			Usage: "Index of primary piece job"},
	},
}

func donePrimaryPieceJob(c *cli.Context) {
	ctx := cliCtx.GetContext()
	if ctx.CurrentService != cliCtx.StoneHubService {
		fmt.Println("please cd StoneHubService namespace, try again")
		return
	}

	txHash, err := hex.DecodeString(c.String("t"))
	if err != nil {
		fmt.Println("tx hash param decode error: ", err)
		return
	}

	// fake the piece checksum
	hash := sha256.New()
	hash.Write([]byte(time.Now().String()))
	checksum := hash.Sum(nil)
	req := &service.StoneHubServiceDonePrimaryPieceJobRequest{
		TxHash: txHash,
		PieceJob: &service.PieceJob{
			StorageProviderSealInfo: &service.StorageProviderSealInfo{
				PieceIdx:          uint32(c.Uint64("i")),
				StorageProviderId: c.String("s"),
				PieceChecksum:     [][]byte{checksum},
			},
		},
	}

	client, err := GetStoneHubClient()
	if err != nil {
		return
	}
	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rsp, err := client.DonePrimaryPieceJob(rpcCtx, req)
	if err != nil {
		fmt.Println("send create object rpc error:", err)
		return
	}
	if rsp.ErrMessage != nil {
		fmt.Println(rsp.ErrMessage)
		return
	}
	fmt.Println("create object success, tx_hash: ", hex.EncodeToString(txHash))
}
