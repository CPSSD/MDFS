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
