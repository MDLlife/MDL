package wallet

import (
	//It's for stringer
	// "github.com/MDLlife/MDL/src/cipher/base58"   
   
	"github.com/MDLlife/MDL/src/cipher"                        
	   
 	 "strings"                                                      
)

const (
// CoinTypeMDL MDL coin type - for future MDL wallet
//CoinTypeMDL wallet.CoinType = "mdlcoin"
)

/*
MDL  address is 3+20+1+4 bytes
- the first 3 bytes are MDL prefix
- the next 25 bytes are the Skycoin-type address
*/

type MDLAddress struct {
	Prefix string              // MDL prefix
	Address cipher.Address     // 25 byte address      PUBKEY HASH + VERSION + CHECKSUM
}

var MdlPrefix = "MDL"

// MDLAddressFromAddress creates MDLAddress from Address
func MDLAddressFromAddress(ad cipher.Address) MDLAddress {
	mdl := MDLAddress{
		Prefix: MdlPrefix,
		Address:ad,
	}
	return mdl
}

//  Create new MDL wallet as Skycoin wallet with MDL address
func NewMDLWallet(wltName string, opts Options) ([]MDLAddress, error)  {
	//  for future MDL wallet
	//opts.Coin = CoinTypeMDL
	 mdlwal, _ := NewWallet(wltName, opts)
	 newMDLaddresses := []MDLAddress{}
	 for i,addr := range mdlwal.GetAddresses() {
	 	newMDLaddresses[i] = MDLAddressFromAddress(addr)
	 }
	 return newMDLaddresses, nil
}

// CutMDL for future MDL wallet
  func CutMDL(str string) string {
        return strings.Replace(str, "MDL", "", 1)
 }
          
          // Stringer for future use (wallet, explorer etc.)
          
// Stringer address as Base58 encoded string
// Returns address as printable version 
//func (addr MDLAddress) String() string {

// cipher Bytes return address as a byte slice
//	return MdlPrefix + string(base58.Hex2Base58(addr.Address.Bytes()))   
//}
