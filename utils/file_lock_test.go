package utils_test

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"github.com/CPSSD/MDFS/utils"
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
		got, err := CheckCrypto()
		if got != c.equal {
			t.Error("Encryption and decryption process failed.\n")
			fmt.Printf("Error: %s\n", err)
		}
	}
}

const chunkSize = 4000

func CheckCrypto() (success bool, err error) {

	usrHome := utils.GetUserHome()

	path := "/path/to/files/"

	keys := usrHome + path + ".private_key_mdfs"
	source1 := usrHome + path + "test.txt"
	encryp1 := usrHome + path + "test.enc"
	result1 := usrHome + path + "result.txt"

	source2 := usrHome + path + "test.jpg"
	encryp2 := usrHome + path + "test.jpg.enc"
	result2 := usrHome + path + "result.jpg"

	utils.GenUserKeys(keys)

	var prk *rsa.PrivateKey
	var puk *rsa.PublicKey

	err = utils.FileToStruct(keys, &prk)
	if err != nil {
		return false, err
	}
	puk = &prk.PublicKey

	user1 := utils.User{Uuid: 1, Pubkey: puk, Privkey: prk}

	//test two files for encryption and then decryption

	// Test 1st file
	err = utils.EncryptFile(source1, encryp1, user1)
	if err != nil {
		return false, err
	}
	err = utils.DecryptFile(encryp1, result1, user1)
	if err != nil {
		return false, err
	}

	test1 := compareFiles(source1, result1)

	// Test 2nd file
	err = utils.EncryptFile(source2, encryp2, user1)
	if err != nil {
		return false, err
	}
	err = utils.DecryptFile(encryp2, result2, user1)
	if err != nil {
		return false, err
	}

	test2 := compareFiles(source2, result2)

	return test1 && test2, err
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
