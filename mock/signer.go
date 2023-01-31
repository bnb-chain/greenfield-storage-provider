package mock

import (
	"time"

	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
)

// SignerServerMock mock signer service
type SignerServerMock struct {
	InscriptionChain *InscriptionChainMock
}

// NewSignerServerMock return SignerServerMock instance
func NewSignerServerMock(chain *InscriptionChainMock) *SignerServerMock {
	return &SignerServerMock{
		InscriptionChain: chain,
	}
}

// BroadcastCreateObjectMessage mock broadcast create object message to inscription chain
func (signer *SignerServerMock) BroadcastCreateObjectMessage(object *ptypesv1pb.ObjectInfo) []byte {
	txHash := hash.GenerateChecksum([]byte(time.Now().String()))
	go func() {
		time.Sleep(1 * time.Second)
		signer.InscriptionChain.CreateObjectByTxHash(txHash, object)
	}()
	return txHash
}

// BroadcastSealObjectMessage mock broadcast seal object message  to inscription chain
func (signer *SignerServerMock) BroadcastSealObjectMessage(object *ptypesv1pb.ObjectInfo) []byte {
	txHash := hash.GenerateChecksum([]byte(time.Now().String()))
	go func() {
		time.Sleep(1 * time.Second)
		signer.InscriptionChain.SealObjectByTxHash(txHash, object)
	}()
	return txHash
}
