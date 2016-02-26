package main

import (
	"crypto/rsa"
	"fmt"
	"github.com/CPSSD/MDFS/utils"
)

func main() {

	var prk rsa.PrivateKey
	var puk rsa.PublicKey

	FileToStruct("/path/to/files/.private_key_mdfs", &prk)
	fmt.Printf("Opened private key file: \n%v\n", prk)

	FileToStruct("/path/to/files/.public_key_mdfs", &puk)
	fmt.Printf("Opened public key file: \n%v\n", puk)
}
