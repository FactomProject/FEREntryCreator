// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

// The main reads the config file, gets values from the command line for the FEREntry,
// and then makes a curl commit and reveal string which it sends to a file.
func main() {
	entry, reveal, targetPriceInDollars, ecAddress, err := CreateFEREntryAndReveal()
	if (err != nil ) {
		fmt.Println("Error: ", err)
		return
	}

	compositionString := GetCurlOutputForComposition(entry, reveal, targetPriceInDollars,ecAddress)

	fmt.Println(compositionString)

	_, err = WriteToFile("FERComposeCurls.dat", compositionString)
	if (err != nil ) {
		fmt.Println("Error: ", err)
		return
	}

	return
}