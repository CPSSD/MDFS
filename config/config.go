package config

import (
	"encoding/json"
	"fmt"
	"os"
)

//TODO maybe should make vars unexported and create accessor methods
type Configuration struct {
	Protocol string
	Host     string
	Port     string
	MdHost   string
	MdPort   string
	Path     string
	Unid     string
}

// saves the passed in configuration to file
func SetConfiguration(conf Configuration, filename string) (err error) {

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

// get configuration information from JSON file
func ParseConfiguration(filename string) Configuration {

	file, _ := os.Open(filename)
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}
