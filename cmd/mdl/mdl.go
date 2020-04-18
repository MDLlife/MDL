/*
mdl daemon
*/
package main

/*
CODE GENERATED AUTOMATICALLY WITH FIBER COIN CREATOR
AVOID EDITING THIS MANUALLY
*/

import (
	"flag"
    "github.com/MDLlife/MDL/src/fiber"
    _ "net/http/pprof"
	"os"

	"github.com/MDLlife/MDL/src/mdl"
	"github.com/MDLlife/MDL/src/readable"
	"github.com/MDLlife/MDL/src/util/logging"
)

var (
	// Version of the node. Can be set by -ldflags
    Version = "0.27.0"
	// Commit ID. Can be set by -ldflags
	Commit = ""
	// Branch name. Can be set by -ldflags
	Branch = ""
	// ConfigMode (possible values are "", "STANDALONE_CLIENT").
	// This is used to change the default configuration.
	// Can be set by -ldflags
	ConfigMode = ""

	logger = logging.MustGetLogger("main")

	// CoinName name of coin
	CoinName = "mdl"

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
		"76.74.178.136:7800",
		"68.183.177.154:7800",
		"128.199.148.6:7800",
	}

	nodeConfig = mdl.NewNodeConfig(ConfigMode, fiber.NodeConfig{
		CoinName:            CoinName,
		GenesisSignatureStr: GenesisSignatureStr,
		GenesisAddressStr:   GenesisAddressStr,
		GenesisCoinVolume:   GenesisCoinVolume,
		GenesisTimestamp:    GenesisTimestamp,
		BlockchainPubkeyStr: BlockchainPubkeyStr,
		BlockchainSeckeyStr: BlockchainSeckeyStr,
		DefaultConnections:  DefaultConnections,
		PeerListURL:         "",
		Port:                7800,
		WebInterfacePort:    8320,
		DataDirectory:       "$HOME/.mdl",

        UnconfirmedBurnFactor:          10,
        UnconfirmedMaxTransactionSize:  32768,
        UnconfirmedMaxDropletPrecision: 3,
        CreateBlockBurnFactor:          10,
        CreateBlockMaxTransactionSize:  32768,
        CreateBlockMaxDropletPrecision: 3,
        MaxBlockTransactionsSize:       32768,

        DisplayName:           "Skycoin",
        Ticker:                "SKY",
        CoinHoursName:         "Coin Hours",
        CoinHoursNameSingular: "Coin Hour",
        CoinHoursTicker:       "SCH",
        ExplorerURL:           "https://explorer.skycoin.com",
        VersionURL:            "https://version.skycoin.com/skycoin/version.txt",
        Bip44Coin:             8000,
	})

	parseFlags = true
)

func init() {
	nodeConfig.RegisterFlags()
}

func main() {
	if parseFlags {
		flag.Parse()
	}

	// create a new fiber coin instance
	coin := mdl.NewCoin(mdl.Config{
		Node: nodeConfig,
		Build: readable.BuildInfo{
			Version: Version,
			Commit:  Commit,
			Branch:  Branch,
		},
	}, logger)

	// parse config values
	if err := coin.ParseConfig(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	// run fiber coin node
	if err := coin.Run(); err != nil {
		os.Exit(1)
	}
}
