package stonehub

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/urfave/cli"

	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"

	types "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"

	cliCtx "github.com/bnb-chain/greenfield-storage-provider/test/test_tool/context"
)

var CreateObjectCommand = cli.Command{
	Name:   "create_object",
	Usage:  "Create object in the stone hub, notice the command is use for testing",
	Action: createObjectToStoneHub,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "w,owner",
			Value: "",
			Usage: "Owner of object"},
		cli.StringFlag{
			Name:  "b,BucketName",
			Value: "",
			Usage: "BucketName of object"},
		cli.StringFlag{
			Name:  "o,ObjectName",
			Value: "",
			Usage: "ObjectName of object"},
		cli.Uint64Flag{
			Name:  "s,Size",
			Value: 0,
			Usage: "Size of object"},
		cli.Uint64Flag{
			Name:  "i,ObjectID",
			Value: 0,
			Usage: "ObjectID(resource id) of object"},
		cli.Uint64Flag{
			Name:  "c,CreateHeight",
			Value: 0,
			Usage: "Height of creating object on the inscription"},
		cli.StringFlag{
			Name:  "sp,SpId",
			Value: "",
			Usage: "Primary storage provider id"},
	},
}

func createObjectToStoneHub(c *cli.Context) {
	ctx := cliCtx.GetContext()
	if ctx.CurrentService != cliCtx.StoneHubService {
		fmt.Println("please cd StoneHubService namespace, try again")
		return
	}

	bucketName := c.String("b")
	if len(bucketName) == 0 {
		fmt.Println("bucket name is empty.")
		return
	}
	objectName := c.String("o")
	if len(objectName) == 0 {
		fmt.Println("object name is empty.")
		return
	}

	// fake the tx hash
	hash := sha256.New()
	hash.Write([]byte(time.Now().String()))
	txHash := hash.Sum(nil)

	object := &types.ObjectInfo{
		Owner:      c.String("w"),
		BucketName: c.String("b"),
		ObjectName: c.String("o"),
		Size:       c.Uint64("s"),
		ObjectId:   c.Uint64("i"),
		Height:     c.Uint64("c"),
		TxHash:     txHash,
		PrimarySp: &types.StorageProviderInfo{
			SpId: c.String("sp"),
		},
	}

	req := &service.StoneHubServiceCreateObjectRequest{
		TxHash:     txHash,
		ObjectInfo: object,
	}

	client, err := GetStoneHubClient()
	if err != nil {
		return
	}
	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rsp, err := client.CreateObject(rpcCtx, req)
	if err != nil {
		fmt.Println("send create object rpc error:", err)
		return
	}
	if rsp.ErrMessage != nil {
		fmt.Println(rsp.ErrMessage)
		return
	}
	fmt.Println("create object success, tx_hash: ", hex.EncodeToString(txHash))
	return
}
