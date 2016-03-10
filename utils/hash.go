package utils

import (
	"bufio"
	"crypto/md5ß"
	"encoding/hex"
	"io"
	"os"
)

// returns the the files md5 checksum in a byte array
// the checksum should only be used if the error returned is nil
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

// used by storage node to find files
func CheckForHash(path, hash string) bool {

	if _, err := os.Stat(path + hash); err == nil {
		return true
	} else {
		return false
	}
}

// network hash functions
// reads bytes from a bufio reader and returns the stirng representation
func ReadHashAsString(r *bufio.Reader) string {

	// make a buffer to hold hash
	buf := make([]byte, 16)
	_, err := r.Read(buf)
	if err != nil && err != io.EOF {
		panic(err)
	}

	hash := hex.EncodeToString(buf)
	return hash
}

// writes a byte array, representing a hash, to a bufio writer
func WriteHash(w *bufio.Writer, hash []byte) error {

	w.Write(hash)
	return w.Flush()
}
