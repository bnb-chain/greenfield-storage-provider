package authenticator

import (
	"encoding/hex"
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type PublicKey = eddsa.PublicKey

// ParsePk will parse eddsa public key from string
func ParsePk(pkStr string) (pk *PublicKey, err error) {
	pkBytes, err := hex.DecodeString(pkStr)
	if err != nil {
		log.Errorf("invalid public key: %+v", err)
		return nil, err
	}
	pk = new(PublicKey)
	size, err := pk.SetBytes(pkBytes)
	if err != nil {
		log.Errorf("invalid public key: %+v", err)
		return nil, err
	}
	if size != 32 {
		log.Errorf("invalid public key")
		return nil, errors.New("invalid public key")
	}
	return pk, nil
}

// Verify will Verify signature of a message with MiMC hash function
func Verify(pk *PublicKey, signature, msg []byte) (bool, error) {
	hasher := mimc.NewMiMC()
	return pk.Verify(signature, msg, hasher)
}

// VerifyEddsaSignature  EDDSA sig verification
func VerifyEddsaSignature(pubKey string, sig, message []byte) error {
	pk, err := ParsePk(pubKey)
	if err != nil {
		log.Errorf("fail to parse public key, pubKey=%s, err=%s", pubKey, err.Error())
		return err
	}
	valid, err := Verify(pk, sig, message)
	if err != nil {
		log.Errorf("fail to Verify signature, sig=%s, message=%s, err=%s", sig, message, err.Error())
		return err
	}
	if !valid {
		return errors.New("invalid signature")
	}
	return nil
}
