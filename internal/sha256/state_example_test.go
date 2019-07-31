//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package sha256

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func ExampleGetHashStateText() {

	// When you need to save the middle state of hash to use sometime in
	// the future, just call GetHashStateText() function simply.
	// When you need this hash with the state, you can call NewHashWithState()
	// function, see the next example. It seems like serialize and deserialize.

	hash := sha256.New()
	if _, err := hash.Write([]byte("something")); err != nil {
		// you should handle the unexpected error
		return
	}
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))
	hashStateText, err := GetHashStateText(hash)
	if err == nil {
		fmt.Println(hashStateText)
	}

	// Output:
	// 3fc9b689459d738f8c88a3a48aa9e33542016b7a4052e001aaa536fca74813cb
	// MP+BAwEBBXN0YXRlAf+CAAEEAQFIAf+EAAEBWAH/hgABAk54AQQAAQNMZW4BBgAAABn/gwEBAQlbOF11aW50MzIB/4QAAQYBEAAAGv+FAQEBCVs2NF11aW50OAH/hgABBgH/gAAAc/+CAQj8agnmZ/y7Z66F/Dxu83L8pU/1OvxRDlJ//JsFaIz8H4PZq/xb4M0ZAUBzb21ldGhpbmcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARIBCQA=
}

func ExampleNewHashWithStateText() {

	// When you saved the middle state of hash, you can recover it at
	// any time.

	stateText := "MP+BAwEBBXN0YXRlAf+CAAEEAQFIAf+EAAEBWAH/hgABAk54AQQAAQNMZW4BBgAAABn/gwEBAQlbOF11aW50MzIB/4QAAQYBEAAAGv+FAQEBCVs2NF11aW50OAH/hgABBgH/gAAAc/+CAQj8agnmZ/y7Z66F/Dxu83L8pU/1OvxRDlJ//JsFaIz8H4PZq/xb4M0ZAUBzb21ldGhpbmcAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARIBCQA="
	hash, _ := NewHashWithStateText(stateText)
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))

	// Output:
	// 3fc9b689459d738f8c88a3a48aa9e33542016b7a4052e001aaa536fca74813cb
}
