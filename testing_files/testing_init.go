package main

import (
	"github.com/CPSSD/MDFS/utils"
	"io"
	"os"
)

func main() {

	err := os.MkdirAll(utils.GetUserHome()+"/.testing_files/", 0777)
	if err != nil {
		panic(err)
	}

	fo, err := os.Create(utils.GetUserHome() + "/.testing_files/test.txt")
	if err != nil {
		panic(err)
	}

	fi, err := os.Open("./testing_files/test.txt")
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(fo, fi)
	if err != nil {
		panic(err)
	}

	fi.Close()
	fo.Close()

	fo, err = os.Create(utils.GetUserHome() + "/.testing_files/test.jpg")
	if err != nil {
		panic(err)
	}

	fi, err = os.Open("./testing_files/test.jpg")
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(fo, fi)
	if err != nil {
		panic(err)
	}

	fi.Close()
	fo.Close()

}
