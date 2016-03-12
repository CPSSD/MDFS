package main

import (
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"os"
)

// for testing setup
func main() {

	err := os.MkdirAll(utils.GetUserHome()+"/.mdservice/", 0700)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(utils.GetUserHome()+"/.mdservice/files/", 0700)
	if err != nil {
		panic(err)
	}

	conf := config.ParseConfiguration("./mdservice/config/mdservice_conf.json")
	conf.Path = utils.GetUserHome() + "/.mdservice/"

	// save the new configuration to file
	err = config.SetConfiguration(conf, utils.GetUserHome()+"/.mdservice/mdservice.json")
	if err != nil {
		panic(err)
	}
}
