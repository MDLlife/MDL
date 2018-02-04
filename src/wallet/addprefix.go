package wallet

import (
"strings"
)

/*
MDL  address is 3+20+1+4 bytes
- the first 3 bytes are MDL prefix
- the next 25 bytes are the Skycoin-type address
*/

// Add MDL  prefix to string address
func AddPrefix(add string) string {
	mdlAddress := "MDL" + add	
	return mdlAddress
}

// CutMDL for future MDL wallet
func CutPrefix(str string) string {
	return strings.Replace(str, "MDL", "", 1)
}
