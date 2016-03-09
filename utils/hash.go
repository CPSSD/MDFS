package utils

import (
	"bufio"
	"io"
	"encoding/hex"
)

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

func WriteHash(w *bufio.Writer, hash []byte) error {

	w.Write(hash)
	return w.Flush()
}