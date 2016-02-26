package utils

import (
	"crypto/rsa"
	"encoding/gob"
	"os"
)

type User struct {
	Uuid    uint64
	Pubkey  *rsa.PublicKey
	Privkey *rsa.PrivateKey
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