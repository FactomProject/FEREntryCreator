
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factom"
	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/goleveldb/leveldb/errors"
	"encoding/json"
	"bytes"
)


// This is a representation of the FER data.  Basically the json of this will be the factom entry content
type FEREntry struct {
	ExpirationHeight       uint32 `json:"expiration_height"`
	TargetActivationHeight uint32 `json:"target_activation_height"`
	Priority               uint32 `json:"priority"`
	TargetPrice            uint64 `json:"target_price"`
	Version                string `json:"version"`
}



func CreateFEREntryAndReveal() (Entry string, Reveal string, targetPriceInDollars float64, err error) {

	// Read the config file
	config, err := readConfigFile("FactomFER.conf")
	if ( err != nil ) {
		errorMessage := errors.New(" Could not find config file FactomFER.conf.\n A sample config file is below, create it if you wish:\n   PaymentPrivateKey = \"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\"\n   SigningPrivateKey = \"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\"\n   Version = \"1.0\"")
		return "", "", 0.0, errorMessage
	}

	// Make an Fer Entry to send along
	theFEREntry := new(FEREntry)
	theFEREntry.Version = config.Version

	// Create and format the payment private key
	var paymentPrivateKey [64]byte
	paymentBytes, err := hex.DecodeString(config.PaymentPrivateKey)
	if (err != nil) {
		return "", "", 0.0, errors.New("Payment private key isn't parsable")
	}
	copy(paymentPrivateKey[:], paymentBytes)
	paymentPublicKey := new([32]byte)
	paymentPublicKey = ed.GetPublicKey(&paymentPrivateKey)

	// Create and format the signing private key
	var signingPrivateKey [64]byte
	signingBytes, err := hex.DecodeString(config.SigningPrivateKey)
	if (err != nil) {
		return "", "", 0.0, errors.New("Signing private key isn't parsable")
	}
	copy(signingPrivateKey[:], signingBytes[:])
	_ = ed.GetPublicKey(&signingPrivateKey)  // Needed to format the public half of the key set


	// Read some values for the FEREntry from the stdIn
	uintValue, err := readStdinUint("Enter the entry expiration height: ", "Bad exipration height", 32)
	if (err != nil) { return }
	theFEREntry.ExpirationHeight = uint32(uintValue)
	uintValue, err = readStdinUint("Enter the target activation height: ", "Bad target activation height", 32)
	if (err != nil) { return }
	theFEREntry.TargetActivationHeight = uint32(uintValue)
	uintValue, err = readStdinUint("Enter the entry priority: ", "Bad priority", 32)
	if (err != nil) { return }
	theFEREntry.Priority = uint32(uintValue)
	uintValue, err = readStdinUint("Enter the new Factoshis Per EC: ", "Bad Factoshis Per EC", 64)
	if (err != nil) { return }
	theFEREntry.TargetPrice = uintValue

	entryJson, err := json.Marshal(theFEREntry)
	if err != nil {
		return "", "", 0.0, errors.New("Could not marshal the data into an FEREntry")
	}
	fmt.Println()

	// Create the factom entry with the signing private key
	signingSignature := ed.Sign(&signingPrivateKey, entryJson)


	// Make a new factom entry and populate it
	e := new(factom.Entry)
	e.ChainID = "eac57815972c504ec5ae3f9e5c1fe12321a3c8c78def62528fb74cf7af5e7389"
	e.ExtIDs = append(e.ExtIDs, signingSignature[:])
	e.Content = entryJson


	// Create the compose and the reveal
	entryCommitJson, err := factom.ComposeEntryCommit(paymentPublicKey, &paymentPrivateKey, e)
	if err != nil { return "", "", 0.0, err }
	revealJson, err := factom.ComposeEntryReveal(e)
	if err != nil { return "", "", 0.0, err }

	impliedFctPrice := float64(0.0)
	if (theFEREntry.TargetPrice != 0) {
		impliedFctPrice = 100000 / float64(theFEREntry.TargetPrice)
	} else {
		return "", "", 0.0, errors.New("Trying to set targetPrice to 0!")
	}

	return string(entryCommitJson), string(revealJson), impliedFctPrice, nil
}



func GetCurlOutputForComposition(entryCommitJson string, revealJson string, targetPriceInDollars float64) (output string){

	var buffer bytes.Buffer

	entry := fmt.Sprintf("    curl -i -X POST -H 'Content-Type: application/json' -d '%s' localhost:8088/v1/commit-entry\n", string(entryCommitJson))
	reveal := fmt.Sprintf("    curl -i -X POST -H 'Content-Type: application/json' -d '%s' localhost:8088/v1/reveal-chain\n", string(revealJson))
	pricePerDollar := fmt.Sprintf("$%.2f", targetPriceInDollars)

	// Make the output file and print to the screen
	buffer.WriteString("***************************************************************************************************\n")
	buffer.WriteString("*\n")
	buffer.WriteString("*   WARNING:  You a making an FERChain entry with the following data:\n")
	buffer.WriteString("*\n")
	buffer.WriteString("*      ")
	buffer.WriteString(entryCommitJson)
	buffer.WriteString("\n")
	buffer.WriteString("*   Implied factoid price:\n")
	buffer.WriteString("*\n")
	buffer.WriteString("*      ")
	buffer.WriteString(pricePerDollar)
	buffer.WriteString("\n")
	buffer.WriteString("***************************************************************************************************\n")
	buffer.WriteString("\n")
	buffer.WriteString( entry)
	buffer.WriteString("\n")
	buffer.WriteString( reveal)
	buffer.WriteString("Done\n")
	buffer.WriteString("\n")

	return buffer.String()
}