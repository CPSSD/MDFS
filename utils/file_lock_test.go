package utils

import (
	"bytes"
	"crypto/rsa"
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

	GenUserKeys()

	var prk *rsa.PrivateKey
	var puk *rsa.PublicKey
	
	FileToStruct("/path/to/files/.private_key_mdfs", &prk)
	FileToStruct("/path/to/files/.public_key_mdfs", &puk)

	user1 := User{Uuid: 1, Pubkey: puk, Privkey: prk}

	//test two files for encryption and then decryption

	// Test 1st file
	source := "/path/to/files/test"
	encryp := "/path/to/files/test.enc"
	result := "/path/to/files/result.txt"

	err := EncryptFile(source, encryp, user1)
	if err != nil {
		return false
	}
	err = DecryptFile(encryp, result, user1)
	if err != nil {
		return false
	}

	test1 := compareFiles(source, result)


	// Test 2nd file
	source = "/path/to/files/david.jpg"
	encryp = "/path/to/files/david.enc"
	result = "/path/to/files/result.jpg"

	err = EncryptFile(source, encryp, user1)
	if err != nil {
		return false
	}
	err = DecryptFile(encryp, result, user1)
	if err != nil {
		return false
	}
	
	test2 := compareFiles(source, result)

	return test1 && test2
}

func compareFiles(file1, file2 string) bool {
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
