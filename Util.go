package main

import (
	"bufio"
	"os"
	"fmt"
	"strings"
	"strconv"
	"github.com/FactomProject/goleveldb/leveldb/errors"
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


func WriteToFile(fileName string, input string) (numberBytes int, err error) {
	f, err := os.Create(fileName)
	if (err != nil) {
		return 0, errors.New("Could not open output file.")
	}
	w := bufio.NewWriter(f)

	defer f.Close()

	numberBytes, err = w.WriteString(input)
	if (err != nil) {
		return 0, errors.New("Could not write to the output file.")
	}

	w.Flush()

	return numberBytes, nil
}