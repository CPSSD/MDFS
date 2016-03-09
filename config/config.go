package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	Protocol string
	Host     string
	Port     string
	Path     string
	Unid     string
}

func SetConfiguration(conf Configuration, filename string) (err error) {

	// saves the passed in configuration to file

	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	encoder := json.NewEncoder(fo)
	err = encoder.Encode(conf)
	if err != nil {
		return err
	}
	return err
}

func ParseConfiguration(filename string) Configuration {
	// get configuration information from JSON file
	file, _ := os.Open(filename)
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}
