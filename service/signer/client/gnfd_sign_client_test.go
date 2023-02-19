package client

import (
	"net"
	"testing"

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

func TestGreenfieldChainSignClient_Sign(t *testing.T) {
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
		GRPCAddr           string
		ChainID            string
		GasLimit           uint64
		OperatorPrivateKey string
		FundingPrivateKey  string
		SealPrivateKey     string
		ApprovalPrivateKey string
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
			name: "Test TestGreenfieldChainSignClient_Sign Case 1",
			fields: fields{
				ChainID:            "greenfield_9000-121",
				GasLimit:           210000,
				GRPCAddr:           "localhost:9090",
				OperatorPrivateKey: "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
				FundingPrivateKey:  "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
				ApprovalPrivateKey: "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
				SealPrivateKey:     "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
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
			// client, err := NewGreenfieldChainClient(tt.fields.config)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("NewGreenfieldChainClient() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }

			// km, err := client.greenfieldClients[tt.args.scope].GetKeyManager()
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("GetKeyManager() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }

			// sig, err := client.Sign(tt.args.scope, tt.args.msg)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("GreenfieldChainClient.Sign() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }
			// TODO: Verify method in greenfield v0.0.5 is incorrect, wait for next release
			// err = stypes.VerifySignature(km.GetAddr(), crypto.Keccak256(tt.args.msg), sig)
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("storage VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }
		})
	}
}
