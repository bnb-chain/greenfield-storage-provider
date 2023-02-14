package signer

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/avast/retry-go/v4"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	ethHd "github.com/evmos/ethermint/crypto/hd"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	RtyAttNum     = uint(5)
	RtyAttem      = retry.Attempts(RtyAttNum)
	RtyDelay      = retry.Delay(time.Millisecond * 500)
	RtyErr        = retry.LastErrorOnly(true)
	RetryInterval = 1 * time.Second

	FallBehindThreshold          = uint64(5)
	SleepIntervalForUpdateClient = 10 * time.Second
	DataSeedDenyServiceThreshold = 60
	RPCTimeout                   = 3 * time.Second
)

type GreenfieldClient struct {
	rpcClient     rpcclient.Client
	txClient      tx.ServiceClient
	authClient    authtypes.QueryClient
	storageClient storagetypes.QueryClient
	Provider      string
	Height        uint64
	UpdatedAt     time.Time
	cdc           *codec.ProtoCodec
}

func (c *GreenfieldClient) GetAccount(address string) (authtypes.AccountI, error) {
	authRes, err := c.authClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}
	var account authtypes.AccountI
	if err := c.cdc.InterfaceRegistry().UnpackAny(authRes.Account, &account); err != nil {
		return nil, err
	}
	return account, nil
}

func newRpcClient(addr string) *http.HTTP {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		panic(err)
	}
	rpcClient, err := http.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		panic(err)
	}
	return rpcClient
}

func grpcConn(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	return conn
}

func initGreenfieldClients(rpcAddrs, grpcAddrs []string) []*GreenfieldClient {
	greenfieldClients := make([]*GreenfieldClient, 0)

	for i := 0; i < len(grpcAddrs); i++ {
		conn := grpcConn(grpcAddrs[i])
		greenfieldClients = append(greenfieldClients, &GreenfieldClient{
			rpcClient:     newRpcClient(rpcAddrs[i]),
			txClient:      tx.NewServiceClient(conn),
			authClient:    authtypes.NewQueryClient(conn),
			storageClient: storagetypes.NewQueryClient(conn),
			Provider:      grpcAddrs[i],
			UpdatedAt:     time.Now(),
			cdc:           cdc(),
		})
	}
	return greenfieldClients
}

func cdc() *codec.ProtoCodec {
	interfaceRegistry := types.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface("AccountI", (*authtypes.AccountI)(nil))
	interfaceRegistry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&authtypes.BaseAccount{},
	)
	interfaceRegistry.RegisterInterface("cosmos.crypto.PubKey", (*cryptotypes.PubKey)(nil))
	interfaceRegistry.RegisterImplementations((*cryptotypes.PubKey)(nil), &ethsecp256k1.PubKey{})
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &storagetypes.MsgSealObject{})
	return codec.NewProtoCodec(interfaceRegistry)
}

func HexToEthSecp256k1PrivKey(hexString string) (*ethsecp256k1.PrivKey, error) {
	bz, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}
	return ethHd.EthSecp256k1.Generate()(bz).(*ethsecp256k1.PrivKey), nil
}
