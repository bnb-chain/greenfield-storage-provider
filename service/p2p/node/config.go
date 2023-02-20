package node

import (
	"os"
	"path/filepath"
	"time"

	dbconf "github.com/bnb-chain/greenfield-storage-provider/store/config"
	tmos "github.com/tendermint/tendermint/libs/os"
)

const (
	// defaultDirPerm is the default permissions used when creating directories.
	defaultDirPerm     = 0700
	defaultConfigDir   = "config"
	DefaultP2PDir      = ".gnfd-sp-p2p"
	defaultNodeKeyName = "node_key.json"
)

var defaultNodeKeyPath = filepath.Join("", defaultNodeKeyName)

var defaultMoniker = getDefaultMoniker()

// getDefaultMoniker returns a default moniker, which is the host name. If runtime
// fails to get the host name, "anonymous" will be returned.
func getDefaultMoniker() string {
	moniker, err := os.Hostname()
	if err != nil {
		moniker = "anonymous"
	}
	return moniker
}

// BaseConfig defines the base configuration for a p2p node
type BaseConfig struct {
	// The root directory for all data.
	// This should be set in viper, so it can unmarshal into this struct
	RootDir string `mapstructure:"home"`

	// A custom human readable name for this node
	Moniker string `mapstructure:"moniker"`

	// A JSON file containing the private key to use for p2p authenticated encryption
	NodeKey string `mapstructure:"node-key-file"`

	// Database directory
	DBPath string `mapstructure:"db-dir"`

	// p2p db, current only support mysql
	DBConfig *dbconf.SqlDBConfig
}

// helper function to make config creation independent of root dir
func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

// NodeKeyFile returns the full path to the node_key.json file
func (cfg BaseConfig) NodeKeyFile() string {
	return rootify(cfg.NodeKey, cfg.RootDir)
}

// DBDir returns the full path to the database directory
func (cfg BaseConfig) DBDir() string {
	return rootify(cfg.DBPath, cfg.RootDir)
}

// EnsureRoot creates the root, config, and data directories if they don't exist,
// and panics if it fails.
func (cfg BaseConfig) EnsureRoot() {
	if err := tmos.EnsureDir(cfg.RootDir, defaultDirPerm); err != nil {
		panic(err.Error())
	}
	if err := tmos.EnsureDir(filepath.Join(cfg.RootDir, defaultConfigDir), defaultDirPerm); err != nil {
		panic(err.Error())
	}
}

// P2PConfig defines the configuration options for the Tendermint peer-to-peer networking layer
type P2PConfig struct { //nolint: maligned
	RootDir string `mapstructure:"home"`

	// Address to listen for incoming connections
	ListenAddress string `mapstructure:"laddr"`

	// Address to advertise to peers for them to dial
	ExternalAddress string `mapstructure:"external-address"`

	// Comma separated list of peers to be added to the peer store
	// on startup. Either BootstrapPeers or PersistentPeers are
	// needed for peer discovery
	BootstrapPeers string `mapstructure:"bootstrap-peers"`

	// Comma separated list of nodes to keep persistent connections to
	PersistentPeers string `mapstructure:"persistent-peers"`

	// UPNP port forwarding
	UPNP bool `mapstructure:"upnp"`

	// MaxConnections defines the maximum number of connected peers (inbound and
	// outbound).
	MaxConnections uint16 `mapstructure:"max-connections"`

	// MaxOutgoingConnections defines the maximum number of connected peers (inbound and
	// outbound).
	MaxOutgoingConnections uint16 `mapstructure:"max-outgoing-connections"`

	// MaxIncomingConnectionAttempts rate limits the number of incoming connection
	// attempts per IP address.
	MaxIncomingConnectionAttempts uint `mapstructure:"max-incoming-connection-attempts"`

	// Set true to enable the peer-exchange reactor
	PexReactor bool `mapstructure:"pex"`

	// Comma separated list of peer IDs to keep private (will not be gossiped to
	// other peers)
	PrivatePeerIDs string `mapstructure:"private-peer-ids"`

	// Time to wait before flushing messages out on the connection
	FlushThrottleTimeout time.Duration `mapstructure:"flush-throttle-timeout"`

	// Maximum size of a message packet payload, in bytes
	MaxPacketMsgPayloadSize int `mapstructure:"max-packet-msg-payload-size"`

	// Rate at which packets can be sent, in bytes/second
	SendRate int64 `mapstructure:"send-rate"`

	// Rate at which packets can be received, in bytes/second
	RecvRate int64 `mapstructure:"recv-rate"`

	// Peer connection configuration.
	HandshakeTimeout time.Duration `mapstructure:"handshake-timeout"`
	DialTimeout      time.Duration `mapstructure:"dial-timeout"`

	// Makes it possible to configure which queue backend the p2p
	// layer uses. Options are: "fifo" and "simple-priority", and "priority",
	// with the default being "simple-priority".
	QueueType string `mapstructure:"queue-type"`
}

// NodeConfig defines the top level configuration for a Metastone node
type NodeConfig struct {
	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`

	P2P *P2PConfig `mapstructure:"p2p"`
}

// DefaultNodeConfig returns a default configuration for a node
func DefaultNodeConfig() NodeConfig {
	return NodeConfig{
		BaseConfig: DefaultBaseConfig(),
		P2P:        DefaultP2PConfig(),
	}
}

// DefaultBaseConfig returns a default base configuration for a node base info
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		RootDir: DefaultP2PDir,
		NodeKey: defaultNodeKeyPath,
		Moniker: defaultMoniker,
		DBPath:  "data",
	}
}

// DefaultP2PConfig returns a default configuration for the peer-to-peer layer
func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		ListenAddress:                 "tcp://127.0.0.1:26656",
		ExternalAddress:               "tcp://127.0.0.1:26676",
		UPNP:                          false,
		MaxConnections:                64,
		MaxOutgoingConnections:        12,
		MaxIncomingConnectionAttempts: 100,
		FlushThrottleTimeout:          100 * time.Millisecond,
		// The MTU (Maximum Transmission Unit) for Ethernet is 1500 bytes.
		// The IP header and the TCP header take up 20 bytes each at least (unless
		// optional header fields are used) and thus the max for (non-Jumbo frame)
		// Ethernet is 1500 - 20 -20 = 1460
		// Source: https://stackoverflow.com/a/3074427/820520
		MaxPacketMsgPayloadSize: 1400,
		SendRate:                5120000, // 5 mB/s
		RecvRate:                5120000, // 5 mB/s
		PexReactor:              true,
		HandshakeTimeout:        20 * time.Second,
		DialTimeout:             3 * time.Second,
		QueueType:               "simple-priority",
	}
}
