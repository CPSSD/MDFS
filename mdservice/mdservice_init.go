package main

import (
	"encoding/json"
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"os"
)

func main() {

	err := os.MkdirAll(utils.GetUserHome()+"/.mdservice/", 0777)
	if err != nil {
		panic(err)
	}

	fo, err := os.Create(utils.GetUserHome() + "/.mdservice/mdservice_conf.json")
	if err != nil {
		panic(err)
	}

	conf := config.ParseConfiguration("./mdservice/mdservice_conf.json")
	conf.Path = utils.GetUserHome() + "/.mdservice/"

	encoder := json.NewEncoder(fo)

	err = encoder.Encode(conf)

	if err != nil {
		panic(err)
	}

	fo.Close()
}
