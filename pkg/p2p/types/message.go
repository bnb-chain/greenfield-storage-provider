package types

import (
	"bytes"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// GetSignBytes returns the ping message bytes to sign over.
func (m *Ping) GetSignBytes() []byte {
	fakeMsg := proto.Clone(m).(*Ping)
	fakeMsg.Signature = []byte{}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

// GetSignBytes returns the pong message bytes to sign over.
func (m *Pong) GetSignBytes() []byte {
	fakeMsg := proto.Clone(m).(*Pong)
	fakeMsg.Signature = []byte{}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

// GetSignBytes returns the get approval request message bytes to sign over.
func (m *GetApprovalRequest) GetSignBytes() []byte {
	object := *m.GetObjectInfo()
	fakeMsg := &GetApprovalRequest{
		SpOperatorAddress: m.GetSpOperatorAddress(),
		ObjectInfo:        &object,
	}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

// GetSignBytes returns the get approval response message bytes to sign over.
func (m *GetApprovalResponse) GetSignBytes() []byte {
	object := *m.GetObjectInfo()
	fakeMsg := &GetApprovalResponse{
		SpOperatorAddress: m.GetSpOperatorAddress(),
		ObjectInfo:        &object,
		ExpiredTime:       m.GetExpiredTime(),
		RefusedReason:     m.GetRefusedReason(),
	}
	fakeMsg.Signature = []byte{}
	bz := ModuleCdc.MustMarshalJSON(fakeMsg)
	return sdk.MustSortJSON(bz)
}

// VerifySignature verifier whether the signer address and signed msg match
func VerifySignature(spOpAddr string, signBytes []byte, sig []byte) error {
	spOpAcc, err := sdk.AccAddressFromHexUnsafe(spOpAddr)
	if err != nil {
		return err
	}
	sigHash := sdk.Keccak256(signBytes)

	if len(sig) != ethcrypto.SignatureLength {
		return errors.Wrapf(sdkerrors.ErrorInvalidSigner, "signature length (actual: %d) doesn't match typical [R||S||V] signature 65 bytes", len(sig))
	}
	if sig[ethcrypto.RecoveryIDOffset] == 27 || sig[ethcrypto.RecoveryIDOffset] == 28 {
		sig[ethcrypto.RecoveryIDOffset] -= 27
	}

	sigPubKeyBytes, err := secp256k1.RecoverPubkey(sigHash, sig)
	if err != nil {
		return errors.Wrap(err, "failed to recover sp operator public key from sig")
	}
	sigPubKey, err := ethcrypto.UnmarshalPubkey(sigPubKeyBytes)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal sp operator public key")
	}
	sigPubKeyAddr := ethcrypto.PubkeyToAddress(*sigPubKey)
	if !bytes.Equal(sigPubKeyAddr.Bytes(), spOpAcc.Bytes()) {
		return errors.Wrapf(sdkerrors.ErrInvalidPubKey, "signer pubkey %s is different from sp operator pubkey %s", sigPubKeyAddr, spOpAcc)
	}

	recoveredSignerAcc := sdk.AccAddress(sigPubKeyAddr.Bytes())
	if !recoveredSignerAcc.Equals(spOpAcc) {
		return errors.Wrapf(sdkerrors.ErrorInvalidSigner, "failed to verify delegated fee payer %s signature", recoveredSignerAcc)
	}

	// VerifySignature of ethsecp256k1 accepts 64 byte signature [R||S]
	// WARNING! Under NO CIRCUMSTANCES try to use pubKey.VerifySignature there
	if !secp256k1.VerifySignature(sigPubKeyBytes, sigHash, sig[:len(sig)-1]) {
		return errors.Wrap(sdkerrors.ErrorInvalidSigner, "unable to verify signer signature of EIP712 typed data")
	}

	return nil
}
