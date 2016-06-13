package main

import (
	"os"
	"fmt"
	"github.com/BurntSushi/toml"
	"reflect"
	"github.com/FactomProject/goleveldb/leveldb/errors"
)



// Info from config file
// Everything that you add to this config struct must be added in the corresponding
// order to the array ConfigFieldNames.  This makes the error messages correct when
// fields are missing
type Config struct {
	PaymentPrivateKey   string
	SigningPrivateKey     string
	Version string
}

func GetConfigFieldName(fieldIndex int) (fieldName string) {
	var ConfigFieldNames = []string {"PaymentPrivateKey", "SigningPrivateKey", "Version"}
	if (fieldIndex >= 0 ) && (fieldIndex <= len(ConfigFieldNames)) { return "" }
	return ConfigFieldNames[fieldIndex]
}



// Reads info from config file.
func readConfigFile(configFileName string) (Config, error) {
	var config Config

	_, err := os.Stat(configFileName)
	if err != nil {

		fmt.Println("sub-error", err)

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
			fmt.Println("Error Couldn't read value for config field ", GetConfigFieldName(i))
		}
	}

	if (fieldsMissed) {
		return config, errors.New("Couldn't read all of config")
	}

	return config, nil
}
