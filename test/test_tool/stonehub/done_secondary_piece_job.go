package stonehub

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/urfave/cli"

	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	cliCtx "github.com/bnb-chain/greenfield-storage-provider/test/test_tool/context"
)

var DoneSecondaryPieceJobCommand = cli.Command{
	Name:   "done_secondary_piece_job",
	Usage:  "Complete secondary piece job",
	Action: doneSecondaryPieceJob,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "t,TxHash",
			Value: "",
			Usage: "Transaction hash of create object"},
		cli.StringFlag{
			Name:  "s,StorageProvider",
			Value: "",
			Usage: "StorageProvider id of secondary"},
		cli.Uint64Flag{
			Name:  "i,PieceIdx",
			Value: 0,
			Usage: "Index of secondary piece job"},
	},
}

func doneSecondaryPieceJob(c *cli.Context) {
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
	var checksums [][]byte
	for i := 0; i < 6; i++ {
		hash := sha256.New()
		hash.Write([]byte(time.Now().String() + string(i)))
		checksum := hash.Sum(nil)
		checksums = append(checksums, checksum)
	}

	req := &service.StoneHubServiceDoneSecondaryPieceJobRequest{
		TxHash: txHash,
		PieceJob: &service.PieceJob{
			StorageProviderSealInfo: &service.StorageProviderSealInfo{
				PieceIdx:          uint32(c.Uint64("i")),
				StorageProviderId: c.String("s"),
				PieceChecksum:     checksums,
			},
		},
	}

	client, err := GetStoneHubClient()
	if err != nil {
		return
	}
	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rsp, err := client.DoneSecondaryPieceJob(rpcCtx, req)
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
