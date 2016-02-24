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
	
	FileToStruct("/path/to/files/.private_key_mdfs", &prk)
	fmt.Printf("Opened private key file: \n%v\n", prk)

	FileToStruct("/path/to/files/.public_key_mdfs", &puk)
	fmt.Printf("Opened public key file: \n%v\n", puk)
}

func StructToFile(e interface{}, filename string) (err error) {
	fileout, err := os.Create(filename)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(fileout)
	encoder.Encode(e)

	fileout.Close()
	return nil
}

func FileToStruct(filename string, e interface{}) (err error) {
	filein, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(filein)
	decoder.Decode(e)

	filein.Close()
	return nil
}