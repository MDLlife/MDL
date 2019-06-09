package mdl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO(therealssj): write better tests
func TestNewParameters(t *testing.T) {
	coinConfig, err := NewParameters("test.fiber.toml", "./testdata")
	require.NoError(t, err)
	require.Equal(t, Parameters{
		Node: NodeParameters{
			GenesisSignatureStr: "eb10468d10054d15f2b6f8946cd46797779aa20a7617ceb4be884189f219bc9a164e56a5b9f7bec392a804ff3740210348d73db77a37adb542a8e08d429ac92700",
			GenesisAddressStr:   "2jBbGxZRGoQG1mqhPBnXnLTxK6oxsTf8os6",
			BlockchainPubkeyStr: "025d096499390a1924969f0991b1e0fd5f37c9ec54f7830f10fa8d911a51bb1e4b",
			BlockchainSeckeyStr: "",
			GenesisTimestamp:    1426562704,
			GenesisCoinVolume:   100e13,
			DefaultConnections: []string{
				"118.178.135.93:6000",
				"47.88.33.156:6000",
				"104.237.142.206:6000",
				"176.58.126.224:6000",
				"172.104.85.6:6000",
				"139.162.7.132:6000",
			},
			Port:                           7800,
			PeerListURL:                    "",
			WebInterfacePort:               8320,
			UnconfirmedBurnFactor:          10,
			UnconfirmedMaxTransactionSize:  777,
			UnconfirmedMaxDropletPrecision: 3,
			CreateBlockBurnFactor:          9,
			CreateBlockMaxTransactionSize:  1234,
			CreateBlockMaxDropletPrecision: 4,
			MaxBlockTransactionsSize: 1111,
		},
		Params: ParamsParameters{
			MaxCoinSupply:              1e8,
			DistributionAddressesTotal: 100,
			InitialUnlockedCount:       25,
			UnlockAddressRate:          5,
			UnlockTimeInterval:         60 * 60 * 24 * 365,
			UserBurnFactor:             3,
			UserMaxTransactionSize:     999,
			UserMaxDropletPrecision:    2,
		},
	}, coinConfig)
}
