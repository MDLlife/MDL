package visor

import "github.com/MDLlife/MDL/src/coin"

const (
	// Maximum supply of skycoins
	MaxCoinSupply uint64 = 1e9 // 100,000,000 million

	// Number of distribution addresses
	DistributionAddressesTotal uint64 = 100

	DistributionAddressInitialBalance uint64 = MaxCoinSupply / DistributionAddressesTotal

	// Initial number of unlocked addresses
	InitialUnlockedCount uint64 = 100

	// Number of addresses to unlock per unlock time interval
	UnlockAddressRate uint64 = 5

	// Unlock time interval, measured in seconds
	// Once the InitialUnlockedCount is exhausted,
	// UnlockAddressRate addresses will be unlocked per UnlockTimeInterval
	UnlockTimeInterval uint64 = 60 * 60 * 24 * 365 // 1 year
)

func init() {
	if MaxCoinSupply%DistributionAddressesTotal != 0 {
		panic("MaxCoinSupply should be perfectly divisible by DistributionAddressesTotal")
	}
}

// Returns a copy of the hardcoded distribution addresses array.
// Each address has 1,000,000 coins. There are 100 addresses.
func GetDistributionAddresses() []string {
	addrs := make([]string, len(distributionAddresses))
	for i := range distributionAddresses {
		addrs[i] = distributionAddresses[i]
	}
	return addrs
}

// Returns distribution addresses that are unlocked, i.e. they have spendable outputs
func GetUnlockedDistributionAddresses() []string {
	// The first InitialUnlockedCount (25) addresses are unlocked by default.
	// Subsequent addresses will be unlocked at a rate of UnlockAddressRate (5) per year,
	// after the InitialUnlockedCount (25) addresses have no remaining balance.
	// The unlock timer will be enabled manually once the
	// InitialUnlockedCount (25) addresses are distributed.

	// NOTE: To have automatic unlocking, transaction verification would have
	// to be handled in visor rather than in coin.Transactions.Visor(), because
	// the coin package is agnostic to the state of the blockchain and cannot reference it.
	// Instead of automatic unlocking, we can hardcode the timestamp at which the first 30%
	// is distributed, then compute the unlocked addresses easily here.

	addrs := make([]string, InitialUnlockedCount)
	for i := range distributionAddresses[:InitialUnlockedCount] {
		addrs[i] = distributionAddresses[i]
	}
	return addrs
}

// Returns distribution addresses that are locked, i.e. they have unspendable outputs
func GetLockedDistributionAddresses() []string {
	// TODO -- once we reach 30% distribution, we can hardcode the
	// initial timestamp for releasing more coins
	addrs := make([]string, DistributionAddressesTotal-InitialUnlockedCount)
	for i := range distributionAddresses[InitialUnlockedCount:] {
		addrs[i] = distributionAddresses[InitialUnlockedCount+uint64(i)]
	}
	return addrs
}

// Returns true if the transaction spends locked outputs
func TransactionIsLocked(inUxs coin.UxArray) bool {
	lockedAddrs := GetLockedDistributionAddresses()
	lockedAddrsMap := make(map[string]struct{})
	for _, a := range lockedAddrs {
		lockedAddrsMap[a] = struct{}{}
	}

	for _, o := range inUxs {
		uxAddr := o.Body.Address.String()
		if _, ok := lockedAddrsMap[uxAddr]; ok {
			return true
		}
	}

	return false
}

var distributionAddresses = [DistributionAddressesTotal]string{
	"2yTcExvfLJZEzvERGmtD6p7Eb1aj5ceJKf",
	"2gEv7WtQEfqf6Br5qSWTLKvPw4Q5AsfcEe8",
	"yBRshTxPszDHB88e9Kg4iLXaefeMMtyeEz",
	"h8tbUcUAUpTgVq3PqMTQcLEH3a6UVJqHmS",
	"2CGrDHn5tSzqzgbkEcNLaTxc6Y4G4fxUt5R",
	"2CucXMcBqWbG7x5Zz5aRKgEbvady2hR8W4T",
	"2VSiY5uuEkaU1oHAH3RvMTowpVJzfyrPHmG",
	"2ZhUg4Y2yk8JZM2fQkEVc3Fbekrcc4SPyrr",
	"6AepEwqm6wCpccEfYA4P5MMZrUviMAfntd",
	"2NkK6MYUud1EonZKQThghxP5PhPsBkpjpfj",
	"Lch6Q2dWFbeHgpbxf2C5vwTe8aiDbZQdyH",
	"2faF2GSNfEN6dLz6uJMvVVPt57JZ379KYY9",
	"bHtBS778Qx8rXGTwnNPCGMdmVqpAedXLo",
	"d8J7GpDSgpnpUP683iSsCqp9xJ5zSq6G1D",
	"LBASagoCNd3T6uEA6NzBdtAVGJAY9tqV3J",
	"RCfPg1KAd3J1RS8uuzWJf2mk6f1fXb3XvE",
	"2QtPsYSsd3kXEVUywJPznowKrWo2NZujGpF",
	"sMmZUQbbyfbHikvZU5rEA2cVrFhtpTmzJe",
	"2BPYaoTUgpVBoKB4x9CmGS9xYmPjVNvNyXJ",
	"VM8d1wv8i8Ekfj3vVuJaV2xW4JyiJXk2eh",
	"4FJa8vDMHR7zHG94ph5rbnFu9t4REKhQv7",
	"28zVeAeKDSXBz8zU2xks4WZQa9uAfCK6t8F",
	"DUM7ByZsEcGkUv4LjcAiNG3qAQPL9UBCKr",
	"zD3HxKSnzEk6PZHY6baUZRybM5V76kn5Yu",
	"jmfywy87ua5zvshGhE9mn1P3jW9gtZhZ3w",
	"2ZQTf1rHMap2ZNqB21RJB8JXcVMaLV7dwCP",
	"2j6krapbWoQ6adjxayrjrMLN3zgPrzQrBjr",
	"2PeCkwEb3AAjCyeYdP3kMettjyYVow7BQ7b",
	"zohhGNXSpeK1zeW7cYxp8qArBJmjBjRNHt",
	"iPDQetuT8ttkJhkhMMvbecYwoL6nthp79N",
	"PqXM33Ca3CpvnmRPrbjMHhaXNPukAwv7xp",
	"C9WDwKP2G1Z4fxmDSgPG6G52UQ9M9iAsU6",
	"2mQdVk2ZovkFJSAUsRQELpz6d8a4xGMzGAD",
	"PyAhPBhDhyiyFmK5iqUwaeaHnfwcHiveeA",
	"2H9p3yWgCkCSHsimhrW2UEEhBVGN6sXFNHj",
	"2QKX3DtYw4RGk8tL4eQHY7GdXKaWSFwqFoU",
	"2W3mcjdHuYCEqih6CwwG5Ry1uztyAZbbxqu",
	"RqeStzP6mAfjUzN1fVhZjk52DFzPr5aqgK",
	"xvHj3JcJwUcEWyMc4LwVW1gPpjKvzDijWa",
	"ax7Aevu7RuEW2kZz3ZaqUtTUmfAT2xscZb",
	"83LkbJwQNihseoE8enovM6sk17co3N8r9J",
	"9ZkH5PhGTkY7hk9rMLKucZuhQj7auUXMCx",
	"gzTdhTy1C7DiHV4Hvrsv6ubGnMR8XTejq8",
	"N7hLmncYEJwFNNCot17Rx1kBfdy4jqMQFg",
	"LqGJQNkqQZwhenqurDPMj3LeECC3YjHh1j",
	"CMpBqPUo5aqa2GKYhZmp5uTD4yoft2Kz2b",
	"25Gjekp2XVUj5hUmdBPMsjFFBj6tUjHdvNo",
	"2jE8YDYntc3Zh49Fer7sN5PbAiSQgDXVcn9",
	"23Vkppq67wXyJyEZEForsst4dgBBNU5Gt4F",
	"2i6g47E8epfAmzdt8Pht7n4YQ1M6B2Vgp6x",
	"9nRRiJ7KgCwph7rp7DnF7FVZeXwf81cRjv",
	"254PHM6M9FU8NvoRKzJgUtnhDDvLsfb2h5e",
	"2gsbE8Pc758Y8XpBFksNcRmZdyqhdTurwD2",
	"21ppdXtZ5XGBqPMHFAmXNDEWEYCuZbik5yg",
	"21kaz29jhcuVesx5TC7UoLbPBZd1mP3u6iv",
	"2G6EkcPB3VVstEbsutpJmr4qj4DyjSg146q",
	"ZGRzoTSpcfMDZTKXtcDKGWzxHHA1mnrkio",
	"UV6Frbq8ZH2MdrN7wpWvDzEESbCRZrw7RT",
	"rCj3yy1VqrCDnEwkWGhb7aoZnY2cPGFCYm",
	"MKydRuNMKPGXBtjUBNfnWy8D3GHFhWgSB3",
	"24xdnmQftXLxpzksEdy5oVta4XySp8Ju6qy",
	"shoqX22yQdLEZD5FHRGwvcsCCqe79dhYZA",
	"L7EPyxqLbZqSNPwB4aoe6VqfmM8eMdkBiC",
	"rHoADCCaXcsh3MLUdRnZLPT8sCh12dfwqH",
	"2HjZj3j6QJsi7PfMHWRJWUzQ5nA7Dc82T2Z",
	"2CoTFajTqzqbKQmF9zh4GgtmSLxxuJ3BMjj",
	"2eYHR2r499JRmGv5NaqXNuv3SD7DPiPQDDN",
	"af3Sysn2TX18GsK2umFpmCqauDbQDiQses",
	"fcippoqW2DdM74SjtHkd46yb9oqTmgAuGV",
	"76pt7UDHHz5fBbKJFsNSJddLxPiG8ceC9X",
	"VcfQS9Y9HFDr74t69PvG2gjxhy27DRp4vq",
	"2Yxfm5j3iXiAp5J1Q7M3wpBj2PZVtY2Zppn",
	"z2pppHAiKxr8iJFZTuo5pMP84Yzxznbp7n",
	"wUFY8boFchDp3rdxRGarLNpYvPzrUzUcwE",
	"297viMSqp7XC66BvQyb3unt34aqKosij7wN",
	"27EpRXxRXzvTEMxPJcCgXHmGDKVwdu5VRCH",
	"2SYPok6qJ34wduR278obKNxq19epNCJWHp",
	"21Z2tgJxz28CuVJnbdu2nPgR6DneyAWAvdp",
	"276NZpiwGst65HqqSnTdkrN6emfU3cGkc7b",
	"2h9ineLfaXSenRTe62MtLEf4HhXEHBkjyTd",
	"G2H5VvPthVqaST1gVkRcRW6ZvM2bfVebV2",
	"2byFQowuUXd8ZSh9Yo5XaGKyCPSCAuY1x4",
	"yUPEkyk9nZBt5cRzLgvcWrWtKaGcrmUt4n",
	"2Hy1FdziwAPoDGX8Y2i8JCuuLW2FLZBVeWi",
	"5HN9iiKWzgW7vWVn4quhqDW4oXNRX8qz4J",
	"Zx5ng6rMhmjDHnTJF1ZDhLaNQSReZuF3wm",
	"KnC2fsvrDWisvRrufL5C4vhjazJRh5R6EV",
	"unikwx3PfPpk8NTUDc2m6SMdfPrGU8jYVw",
	"i5e1FhB4LZZZ6JMU7Asb1JQrJaAs6TX3bZ",
	"VYw495Nio7NBsi33yZhLgg6TDqqoT2dpxC",
	"2XRiQDtVD8Wwxzfkc9QmepqcUkHaWo1C1ya",
	"2gQpmhe58LsMRCEU9Tjx4e2DQDJbjVHt8AD",
	"2KBKLSxxbEfLsZVaz5JvJdk5inTakw6AeFj",
	"2K8Mh6ecE27nWQLXWxB1JT13BZT8jAEzKn9",
	"KgVwS2ZCfNwuSss55jYXuizgUJrxrzjA1G",
	"EVa63R7z35VMWJwEoMb7eMXpGVNKApbFTM",
	"9kkcCzNyGv7pq7AcejQpLRxVCE8bQqZkDV",
	"FiMPcfLvYAqkGeh1geL5SgxQQdHWyqNFxV",
	"2Z4vomEgh5GWrtHB4NSSr29Xwnq1iSKZBi9",
	"2m1ma8h29fr6d64vK33xNWaQANiuKhvb7np",
}
