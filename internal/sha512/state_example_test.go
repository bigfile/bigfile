//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package sha512

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

func ExampleGetHashStateText() {

	// When you need to save the middle state of hash to use sometime in
	// the future, just call GetHashStateText() function simply.
	// When you need this hash with the state, you can call NewHashWithState()
	// function, see the next example. It seems like serialize and deserialize.

	hash := sha512.New()
	if _, err := hash.Write([]byte("hello world")); err != nil {
		// you should handle the unexpected error
		return
	}
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))
	hashStateText, err := GetHashStateText(hash)
	if err == nil {
		fmt.Println(hashStateText)
	}

	// Output:
	// 309ecc489c12d6eb4cc40f50c902f2b4d0ed77ee511a7c7a9bcd3ca86d4cd86f989dd35bc5ff499670da34255b45b0cfd830e81f605dcf7dc5542e93ae9cd76f
	// MP+BAwEBBVN0YXRlAf+CAAEEAQFIAf+EAAEBWAH/hgABAk54AQQAAQNMZW4BBgAAABn/gwEBAQlbOF11aW50NjQB/4QAAQYBEAAAHP+FAQEBClsxMjhddWludDgB/4YAAQYB/gEAAAD/1P+CAQj4agnmZ/O8yQj4u2euhYTKpzv4PG7zcv6U+Cv4pU/1Ol8dNvH4UQ5Sf63mgtH4mwVojCs+bB/4H4PZq/tBvWv4W+DNGRN+IXkB/4BoZWxsbyB3b3JsZAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEWAQsA
}

func ExampleNewHashWithStateText() {

	// When you saved the middle state of hash, you can recover it at
	// any time.

	stateText := "MP+BAwEBBVN0YXRlAf+CAAEEAQFIAf+EAAEBWAH/hgABAk54AQQAAQNMZW4BBgAAABn/gwEBAQlbOF11aW50NjQB/4QAAQYBEAAAHP+FAQEBClsxMjhddWludDgB/4YAAQYB/gEAAAD/1P+CAQj4agnmZ/O8yQj4u2euhYTKpzv4PG7zcv6U+Cv4pU/1Ol8dNvH4UQ5Sf63mgtH4mwVojCs+bB/4H4PZq/tBvWv4W+DNGRN+IXkB/4BoZWxsbyB3b3JsZAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEWAQsA"
	hash, _ := NewHashWithStateText(stateText)
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))

	// Output:
	// 309ecc489c12d6eb4cc40f50c902f2b4d0ed77ee511a7c7a9bcd3ca86d4cd86f989dd35bc5ff499670da34255b45b0cfd830e81f605dcf7dc5542e93ae9cd76f
}
