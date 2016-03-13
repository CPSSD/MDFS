package main

import (
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"os"
)

// for testing setup
func main() {

	err := os.MkdirAll(utils.GetUserHome()+"/.mdfs/stnode/", 0700)
	if err != nil {
		panic(err)
	}

	conf := config.ParseConfiguration("./storagenode/config/stnode_conf.json")
	conf.Path = utils.GetUserHome() + "/.mdfs/stnode/"

	// save the new configuration to file
	err = config.SetConfiguration(conf, conf.Path+".stnode_conf.json")
	if err != nil {
		panic(err)
	}
}
