package utils

import (
	"crypto/md5"
	"io"
	"os"
)

func ComputeMd5(filepath string) ([]byte, error) {

	// byte array to hold md5 sum
	var result []byte
	file, err := os.Open(filepath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := md5.New() // type hash.Hash
	// copy the contents of the file into hash
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}
