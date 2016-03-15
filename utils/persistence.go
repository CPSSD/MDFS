package utils

import (
	"crypto/rsa"
	"encoding/gob"
	"os"
	"path"
	"strings"
)

type User struct {
	Uuid    uint64
	Uname   string
	Pubkey  *rsa.PublicKey
	Privkey *rsa.PrivateKey
}

type Stnode struct {
	Unid     string
	Protocol string
	NAddress string
}

type Group struct {
	Gid     uint64 //gid
	Gname   string
	Owner   uint64   //uuids
	Members []uint64 //uuids
}

type FileDesc struct {
	Protected   bool
	Owner       uint64
	Permissions uint16
	Hash        string
	Stnode      string
}

func IsHidden(filepath string) (hidden bool) {
	return strings.HasPrefix(path.Base(filepath), ".")
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
