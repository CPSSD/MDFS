package utils

import (
	"bufio"
	"io"
	"net"
	"os"
)

// serious refactorisation required here
// both functions almost identical
func ReceiveFile(conn net.Conn, r *bufio.Reader, filepath string) {

	// open output file
	fo, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}

	//close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// create a write buffer for the output file
	w := bufio.NewWriter(fo)

	// make a buffer to hold read chunks
	// find a good length
	buf := make([]byte, 16)

	for {

		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}
}

func SendFile(conn net.Conn, w *bufio.Writer, filepath string) {

	// open input file
	fi, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}

	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	// create read buffer for the file
	r := bufio.NewReader(fi)

	// make a buffer to hold read chunks
	// find a good length
	buf := make([]byte, 16)

	for {

		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}
}
