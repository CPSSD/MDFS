package utils

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"
)

// needs to be fixed
func TestEncryption(t *testing.T) {
	var tests = []struct {
		equal bool
	}{
		{true},
	}
	for _, c := range tests {
		got := CheckFiles()
		if got != c.equal {
			t.Error("Encryption and decryption process failed.")
		}
	}
}

const chunkSize = 4000

func CheckFiles() bool {

	//test two files for encryption and then decryption
	source := "/path/to/files/test"
	encryp := "/path/to/files/test.enc"
	result := "/path/to/files/result.txt"

	EncryptFile(source, encryp)
	DecryptFile(encryp, result)

	test1 := CompareFiles(source, result)

	source = "/path/to/files/david.jpg"
	encryp = "/path/to/files/david.enc"
	result = "/path/to/files/result.jpg"

	EncryptFile(source, encryp)
	DecryptFile(encryp, result)

	test2 := CompareFiles(source, result)

	return test1 && test2
}

func CompareFiles(file1, file2 string) bool {
	// Check file size ...

	f1, err := os.Open(file1)
	if err != nil {
		log.Fatal(err)
	}

	f2, err := os.Open(file2)
	if err != nil {
		log.Fatal(err)
	}

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				log.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}
