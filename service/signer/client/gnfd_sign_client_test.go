package client

import (
	"testing"

	"github.com/bnb-chain/greenfield/sdk/keys"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestGreenfieldChainSignClient_Sign(t *testing.T) {
	type fields struct {
		ApprovalPrivateKey string
	}
	type args struct {
		msg []byte
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
				ApprovalPrivateKey: "d710d9e03466d6236d1ac2e70712b1e2ed7324b1d7f233f8887d3a703626fb9f",
			},
			args: args{
				msg: []byte("hello world"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km, err := keys.NewPrivateKeyManager(tt.fields.ApprovalPrivateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeyManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sig, err := km.Sign(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("GreenfieldChainClient.Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = storagetypes.VerifySignature(km.GetAddr(), crypto.Keccak256(tt.args.msg), sig)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
