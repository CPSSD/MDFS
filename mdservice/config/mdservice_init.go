package mdservice

import (
	"encoding/json"
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"os"
)

func Setup(path string, port string) error {

	// create the supplied file path
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return err
	}

	// create a subdirectory called files
	err = os.MkdirAll(path+"files/", 0700)
	if err != nil {
		return err
	}

	// get the sample configuration file from the repo
	conf := config.ParseConfiguration("./mdservice/config/mdservice_conf.json")
	conf.Path = path // change the path variable
	conf.Port = port // change the port variable

	// encode the new object to a json file
	err = config.SetConfiguration(conf, path+"stnode_conf.json")
	return err
}
