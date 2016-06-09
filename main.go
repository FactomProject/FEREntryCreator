// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	ed "github.com/FactomProject/ed25519"
	"bufio"
	"fmt"
	"github.com/FactomProject/factom"
	"os"
	"strconv"
	"strings"
	"github.com/BurntSushi/toml"
	"reflect"
	"github.com/FactomProject/goleveldb/leveldb/errors"
	"encoding/hex"
)


// This function creates a simple buffer input and reads a value from teh command line.
func readStdinUint(prompt string, errorMessage string, intSize int) (uint64, error) {

	reader := bufio.NewReader(os.Stdin)

	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(errorMessage, "      read = ", text, "    err = ", err)
		return 0, errors.New(fmt.Sprintf(errorMessage, "  Error: %s", err))
	}
	text = strings.Replace(text, "\n", "", -1)
	uintValue, err := strconv.ParseUint(text, 10, intSize)
	if err != nil {
		fmt.Println(errorMessage, "      read = ", text, "    err = ", err)
		return 0, errors.New(fmt.Sprintf(errorMessage, "  Error: %s", err))
	}

	return uintValue, nil
}


// Reads info from config file.
func readConfigFile(configFileName string) (Config, error) {
	var config Config

	_, err := os.Stat(configFileName)
	if err != nil {
		return config, errors.New(fmt.Sprintf("Config file is missing: %s", configFileName))
	}

	if _, err := toml.DecodeFile(configFileName, &config); err != nil {
		return config, errors.New(fmt.Sprintf("Error reading your file: %s  Error: %s", configFileName, err))
	}

	fieldsMissed := false;
	fields := reflect.ValueOf(config)
	values := make([]interface{}, fields.NumField())
	for i := 0; i < fields.NumField(); i++ {
		values[i] = fields.Field(i).Interface()
		if (values[i] == "" ) {
			fieldsMissed = true;
			fmt.Println("Error Couldn't read value for config field ", ConfigFieldNames[i])
		}
	}

	if (fieldsMissed) {
		return config, errors.New("Couldn't read all of config")
	}

	return config, nil
}

// This is a representation of the FER data.  Basically the json of this will be the factom entry content
type FEREntry struct {
	ExpirationHeight       uint32 `json:"exiration_height"`
	TargetActivationHeight uint32 `json:"target_activation_height"`
	Priority               uint32 `json:"priority"`
	TargetPrice            uint64 `json:"target_price"`
	Version                string `json:"version"`
}


// Info from config file
// Everything that you add to this config struct must be added in the corresponding
// order to the array ConfigFieldNames.  This makes the error messages correct when
// fields are missing
type Config struct {
	PaymentPrivateKey   string
	SigningPrivateKey     string
	Version string
}
var ConfigFieldNames = []string {"PaymentPrivateKey", "SigningPrivateKey", "Version"}


// THe main reads the config file, gets values from the command line for the FERENtry,
// and then makes a curl commit and reveal string which it sends to a file.
func main() {

	// Read the config file
	config, err := readConfigFile("FactomFER.conf")
	if ( err != nil ) {

		fmt.Println("Error: ", err)
		return
	}

	// Make an Fer Entry to send along
	theFEREntry := new(FEREntry)
	theFEREntry.Version = config.Version


	// Create and format the payment private key
	var paymentPrivateKey [64]byte
	paymentBytes, err := hex.DecodeString(config.PaymentPrivateKey)
	if (err != nil) {
		fmt.Println("Payment private key isn't parsable")
		return
	}
	copy(paymentPrivateKey[:], paymentBytes)
	paymentPublicKey := new([32]byte)
	paymentPublicKey = ed.GetPublicKey(&paymentPrivateKey)

	// Create and format the signing private key
	var signingPrivateKey [64]byte
	signingBytes, err := hex.DecodeString(config.SigningPrivateKey)
	if (err != nil) {
		fmt.Println("Signing private key isn't parsable")
		return
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
		fmt.Println("Could not marshal the data into an FEREntry")
		return
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
	composeJson, err := factom.ComposeEntryCommit(paymentPublicKey, &paymentPrivateKey, e)
	if err != nil { return }
	revealJson, err := factom.ComposeEntryReveal(e)
	if err != nil { return }


	// Make the output file and print to the screen
	fmt.Println("***************************************************************************************************")
	fmt.Println()
	fmt.Println("   WARNING:  You a making an FERChain entry with the following data:")
	fmt.Println()
	fmt.Println("      ", string(entryJson))
	fmt.Println()
	fmt.Println("   Implied factoid price:")
	fmt.Println()
	if (float64(theFEREntry.TargetPrice) != 0) {
		impliedFctPrice := 100000 / float64(theFEREntry.TargetPrice)
		fmt.Printf( "       $%.2f\n", impliedFctPrice)
	} else {
		fmt.Println("       !!!!!!!!!!!!!!!!!!!!!!!!!  IMPLIED PRICE of INFINITE   Aborting!\n")
	}
	fmt.Println()
	fmt.Println("***************************************************************************************************")
	fmt.Println()
	entry := fmt.Sprintf("    curl -i -X POST -H 'Content-Type: application/json' -d '%s' localhost:8088/v1/commit-entry", string(composeJson))
	fmt.Println(entry, "\n")
	reveal := fmt.Sprintf("    curl -i -X POST -H 'Content-Type: application/json' -d '%s' localhost:8088/v1/reveal-chain", string(revealJson))
	fmt.Println( reveal, "\n")
	fmt.Println("Done")
}