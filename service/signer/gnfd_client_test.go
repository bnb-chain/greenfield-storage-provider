package signer

import (
	"net"
	"testing"

	stypes "github.com/bnb-chain/greenfield/x/storage/types"
	"google.golang.org/grpc"
)

func MockGrpcServer() error {
	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	return server.Serve(l)
}

func TestGreenfieldChainClient_Sign(t *testing.T) {
	stop := make(chan error, 1)
	go func() {
		if err := MockGrpcServer(); err != nil {
			stop <- err
		}
	}()
	select {
	case err := <-stop:
		t.Fatal(err)
	default:
	}
	type fields struct {
		config *GreenfieldChainConfig
	}
	type args struct {
		scope SignType
		msg   []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test TestGreenfieldChainClient_Sign Case 1",
			fields: fields{
				config: &GreenfieldChainConfig{
					ChainId:            9000,
					GasLimit:           210000,
					GRPCAddr:           "localhost:9090",
					ChainIdString:      "greenfield_9000-121",
					OperatorPrivateKey: "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
					FundingPrivateKey:  "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
					ApprovalPrivateKey: "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
					SealPrivateKey:     "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
				},
			},
			args: args{
				scope: SignApproval,
				msg:   []byte("hello world"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGreenfieldChainClient(tt.fields.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGreenfieldChainClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			km, err := client.greenfieldClients[tt.args.scope].GetKeyManager()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeyManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sig, err := client.Sign(tt.args.scope, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("GreenfieldChainClient.Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// TODO: verify failed, need to confirm
			err = stypes.VerifySignature(km.GetAddr(), tt.args.msg, sig)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
