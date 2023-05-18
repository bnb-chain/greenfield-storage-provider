package p2pnode

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	// PingPeriodMin defines the min value that the period of sending ping request
	PingPeriodMin = 1
	// DailTimeout defines default value that the timeout of dail other p2p node
	DailTimeout = 1
	// DefaultDataPath defines default value that the path of peer store
	DefaultDataPath = "./data"
	// P2PNode defines the p2p protocol node name
	P2PNode                           = "p2p_protocol_node"
	MinSecondaryApprovalExpiredHeight = 900
)

// MakeMultiaddr new multi addr by address
func MakeMultiaddr(address string) (ma.Multiaddr, error) {
	addrInfo := strings.Split(strings.TrimSpace(address), ":")
	if len(addrInfo) != 2 {
		err := fmt.Errorf("failed to parser address '%s' configuration", address)
		return nil, err
	}
	host := strings.TrimSpace(addrInfo[0])
	if net.ParseIP(host) == nil {
		hosts, err := net.LookupHost(host)
		if err != nil {
			err = fmt.Errorf("failed to parse address '%s' domain", address)
			return nil, err
		}
		if len(hosts) == 0 {
			err = fmt.Errorf("failed to parse address '%s' domain, the empty ip list", address)
			return nil, err
		}
		// addr corresponds to node id one by one, only use the first
		for _, h := range hosts {
			// TODO:: support IPv6
			if strings.Contains(h, ":") {
				continue
			}
			host = h
			break
		}
		if net.ParseIP(host) == nil {
			err = fmt.Errorf("failed to parse address '%s' domain, no usable ip", address)
			return nil, err
		}
	}
	port, err := strconv.Atoi(strings.TrimSpace(addrInfo[1]))
	if err != nil {
		return nil, err
	}
	log.Infow("parser p2p node address", "ip", host, "port", port)
	return ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", host, port))
}

// MakeBootstrapMultiaddr new bootstrap nodes multi addr
func MakeBootstrapMultiaddr(bootstrap []string) (peersIDs []peer.ID, addrs []ma.Multiaddr, err error) {
	for _, boot := range bootstrap {
		boot = strings.TrimSpace(boot)
		bootInfo := strings.Split(boot, "@")
		if len(bootInfo) != 2 {
			err = fmt.Errorf("failed to parser bootstrap '%s' configuration", boot)
			return nil, nil, err
		}
		bootID, err := peer.Decode(strings.TrimSpace(bootInfo[0]))
		if err != nil {
			log.Errorw("failed to decode bootstrap id", "bootstrap", boot, "error", err)
			return nil, nil, err
		}
		addr, err := MakeMultiaddr(strings.TrimSpace(bootInfo[1]))
		if err != nil {
			log.Errorw("failed to make bootstrap multi addr", "bootstrap", boot, "error", err)
			return nil, nil, err
		}
		peersIDs = append(peersIDs, bootID)
		addrs = append(addrs, addr)
	}
	return peersIDs, addrs, err
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
