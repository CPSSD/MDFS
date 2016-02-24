package main

import (
	"crypto/rsa"
	"encoding/gob"
	"fmt"
	"os"
)

func main() {

	var prk rsa.PrivateKey
	var puk rsa.PublicKey

	prf, err := os.Open("/path/to/files/.private_key_mdfs")
	if err != nil {
		panic(err)
	}
	defer prf.Close()

	puf, err := os.Open("/path/to/files/.public_key_mdfs")
	if err != nil {
		panic(err)
	}
	defer puf.Close()

	var decoder gob.Encoder

	decoder = gob.NewDecoder(prf)
	decoder.Decode(&prk)
	fmt.Printf("Opened private key file: \n%v\n", prk)

	decoder = gob.NewDecoder(puf)
	decoder.Decode(&puk)
	fmt.Printf("Opened public key file: \n%v\n", puk)
	// code duplication
}
