package main

import (
	_ "net/http/pprof"

	"github.com/MDLlife/MDL/src/skycoin"
	"github.com/MDLlife/MDL/src/util/logging"
	"github.com/MDLlife/MDL/src/visor"
)

var (
	// Version of the node. Can be set by -ldflags
	Version = "0.24.1"
	// Commit ID. Can be set by -ldflags
	Commit = ""
	// Branch name. Can be set by -ldflags
	Branch = ""
	// ConfigMode (possible values are "", "STANDALONE_CLIENT").
	// This is used to change the default configuration.
	// Can be set by -ldflags
	ConfigMode = ""

	help = false

	logger = logging.MustGetLogger("main")

	// GenesisSignatureStr hex string of genesis signature
	GenesisSignatureStr = "97f68c5564c8526a77a26c54e48c005c18ee76a92a7d0ee397a2e3bd25e5c74a1630952716f3281362f8b2baf22139282ab6b2f3d0e5ee825a69690e76d4401e00"
	// GenesisAddressStr genesis address string
	GenesisAddressStr = "zVebXKCqbtGJMnEoGdFkewvUN5KMryrRTc"
	// BlockchainPubkeyStr pubic key string
	BlockchainPubkeyStr = "025d096499390a1924969f0991b1e0fd5f37c9ec54f7830f10fa8d911a51bb1e4b"
	// BlockchainSeckeyStr empty private key string
	BlockchainSeckeyStr = ""

	// GenesisTimestamp genesis block create unix time
	GenesisTimestamp uint64 = 1516848705
	// GenesisCoinVolume represents the coin capacity
	GenesisCoinVolume uint64 = 1000e12

	// DefaultConnections the default trust node addresses
	DefaultConnections = []string{
		"208.110.84.122:7800",
		"69.90.132.231:7800",
		"76.74.178.136:7800",
		"64.34.218.31:7800"}
)

func main() {
	// get node config
	nodeConfig := skycoin.NewNodeConfig(ConfigMode, skycoin.NodeParameters{
		GenesisSignatureStr: GenesisSignatureStr,
		GenesisAddressStr:   GenesisAddressStr,
		GenesisCoinVolume:   GenesisCoinVolume,
		GenesisTimestamp:    GenesisTimestamp,
		BlockchainPubkeyStr: BlockchainPubkeyStr,
		BlockchainSeckeyStr: BlockchainSeckeyStr,
		DefaultConnections:  DefaultConnections,
		PeerListURL:         "https://downloads.skycoin.net/blockchain/peers.txt",
		Port:                7800,
		WebInterfacePort:    8320,
		DataDirectory:       "$HOME/.mdl",
		ProfileCPUFile:      "mdl.prof",
	})

	// create a new fiber coin instance
	coin := skycoin.NewCoin(
		skycoin.Config{
			Node: *nodeConfig,
			Build: visor.BuildInfo{
				Version: Version,
				Commit:  Commit,
				Branch:  Branch,
			},
		},
		logger,
	)

	// parse config values
	coin.ParseConfig()

	// run fiber coin node
	coin.Run()
}
