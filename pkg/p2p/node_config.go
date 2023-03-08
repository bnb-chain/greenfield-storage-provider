package p2p

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	// PingPeriodMin defines the min value that the period of sending ping request
	PingPeriodMin = 1
	// DailTimeOut defines default value that the timeout of dail other p2p node
	DailTimeOut = 1
	// DefaultDataPath defines default value that the path of peer store
	DefaultDataPath = "./data"
	// P2PNode defines the p2p protocol node name
	P2PNode = "p2p_protocol_node"
)

// NodeConfig defines the p2p node config
type NodeConfig struct {
	// Address defines the p2p listen address, used to make multiaddr
	Address string
	// SpOperatorAddress defines the sp operator public key
	SpOperatorAddress string
	// PrivKey defines the p2p node private key, only support Secp256k1
	// if default value, creates random private-public key pairs
	// TODO::support more crypto algorithm for generating and verifying private-public key pairs
	PrivKey string
	// Bootstrap defines Bootstrap p2p node, cannot be empty
	// format: [node_id1@host1:port1, node_id2@host2:port2]
	Bootstrap []string
	// PingPeriod defines the period of ping other p2p nodes
	PingPeriod int
}

// ParseConfing parsers the configuration into a format that go-libp2p can use
func (cfg *NodeConfig) ParseConfing() (privKey crypto.PrivKey, hostAddr ma.Multiaddr, bootstrapIDs []peer.ID, bootstrapAddrs []ma.Multiaddr, err error) {
	if len(cfg.PrivKey) > 0 {
		priKeyBytes, err := hex.DecodeString(cfg.PrivKey)
		if err != nil {
			log.Errorw("failed to hex decode private key", "priv_key", cfg.PrivKey, "error", err)
			return privKey, hostAddr, bootstrapIDs, bootstrapAddrs, err
		}
		privKey, err = crypto.UnmarshalSecp256k1PrivateKey(priKeyBytes)
		if err != nil {
			log.Errorw("failed to unmarshal secp256k1 private key", "priv_key", cfg.PrivKey, "error", err)
			return privKey, hostAddr, bootstrapIDs, bootstrapAddrs, err
		}
	} else {
		privKey, _, err = crypto.GenerateKeyPair(crypto.Secp256k1, 256)
		if err != nil {
			log.Errorw("failed to generate secp256k1 key pair", "error", err)
			return privKey, hostAddr, bootstrapIDs, bootstrapAddrs, err
		}
	}

	addrInfo := strings.Split(strings.TrimSpace(cfg.Address), ":")
	if len(addrInfo) != 2 {
		err = fmt.Errorf("failed to parser p2p listen address '%s' configuration", cfg.Address)
		return privKey, hostAddr, bootstrapIDs, bootstrapAddrs, err
	}
	host := strings.TrimSpace(addrInfo[0])
	port, err := strconv.Atoi(strings.TrimSpace(addrInfo[1]))
	if err != nil {
		return privKey, hostAddr, bootstrapIDs, bootstrapAddrs, err
	}
	hostAddr, err = MakeMultiaddr(host, port)
	if err != nil {
		log.Errorw("failed to make local mulit addr", "address", cfg.Address, "error", err)
		return privKey, hostAddr, bootstrapIDs, bootstrapAddrs, err
	}
	bootstrapIDs, bootstrapAddrs, err = MakeBootstrapMultiaddr(cfg.Bootstrap)
	if err != nil {
		return privKey, hostAddr, bootstrapIDs, bootstrapAddrs, err
	}
	return privKey, hostAddr, bootstrapIDs, bootstrapAddrs, err
}

// MakeMultiaddr new multi addr by address
func MakeMultiaddr(host string, port int) (ma.Multiaddr, error) {
	return ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", host, port))
}

// MakeBootstrapMultiaddr new bootstrap nodes multi addr
func MakeBootstrapMultiaddr(bootstrap []string) (peersID []peer.ID, addrs []ma.Multiaddr, err error) {
	for _, boot := range bootstrap {
		boot = strings.TrimSpace(boot)
		bootInfo := strings.Split(boot, "@")
		if len(bootInfo) != 2 {
			err = fmt.Errorf("failed to parser bootstrap '%s' configuration", boot)
			return peersID, addrs, err
		}
		bootID, err := peer.Decode(strings.TrimSpace(bootInfo[0]))
		if err != nil {
			log.Errorw("failed to decode bootstrap id", "bootstrap", boot, "error", err)
			return peersID, addrs, err
		}
		addrInfo := strings.Split(strings.TrimSpace(bootInfo[1]), ":")
		if len(addrInfo) != 2 {
			err = fmt.Errorf("failed to parser bootstrap '%s' configuration", boot)
			return peersID, addrs, err
		}
		host := strings.TrimSpace(addrInfo[0])
		port, err := strconv.Atoi(strings.TrimSpace(addrInfo[1]))
		if err != nil {
			return peersID, addrs, err
		}
		addr, err := MakeMultiaddr(host, port)
		if err != nil {
			log.Errorw("failed to make bootstrap multi addr", "bootstrap", boot, "error", err)
			return peersID, addrs, err
		}
		peersID = append(peersID, bootID)
		addrs = append(addrs, addr)
	}
	return peersID, addrs, err
}
