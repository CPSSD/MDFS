package main

import (
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"os"
)

func main() {

	err := os.MkdirAll(utils.GetUserHome()+"/.stnode/", 0700)
	if err != nil {
		panic(err)
	}

	conf := config.ParseConfiguration("./storagenode/config/stnode_conf.json")
	conf.Path = utils.GetUserHome() + "/.stnode/"

	// save the new configuration to file
	err = config.SetConfiguration(conf, utils.GetUserHome()+"/.stnode/stnode_conf.json")
	if err != nil {
		panic(err)
	}
}
