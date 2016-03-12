package storagenode

import (
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"os"
)

func Setup(path string, port string) error {

	err := os.MkdirAll(path, 0700)
	if err != nil {
		return err
	}

	// get the sample configuration from the repo
	conf := config.ParseConfiguration("./storagenode/config/stnode_conf.json")
	conf.Path = path // change the path variable
	conf.Port = port // change the path variable

	// save the new configuration to file
	err = config.SetConfiguration(conf, path+"stnode_conf.json")
	return err
}
