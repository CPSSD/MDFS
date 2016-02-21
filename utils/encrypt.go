package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"crypto/rsa"
	"encoding/gob"
	"errors"
)

var (
	ErrPrivExists = errors.New("privatekey exists, but no publickey")
	ErrPublExists = errors.New("publickey exists, but no privatekey")
	ErrKeyPairExists = errors.New("both privatekey and publickey exist")
)

func GenUserKeys() (success bool, err error) {
	// generate a user's public and private key
	// should only be called if they do not exist already


	if _, err := os.Stat("/path/to/files/.private_key_mdfs"); err == nil {
  		
  		// /path/to/files/.private_key_mdfs exists
		if _, err := os.Stat("/path/to/files/.public_key_mdfs"); err == nil {

			return false, ErrKeyPairExists
		}
		return false, ErrPrivExists
	}
	if _, err := os.Stat("/path/to/files/.public_key_mdfs"); err == nil {
  		// path/to/whatever exists
		return false, ErrPublExists
	}

	// generate a new RSA key
	/*if privatekey, err := rsa.GenerateKey(rand.Reader, 1024); err != nil {
		panic(err)
	}*/

	privatekey, err := rsa.GenerateKey(rand.Reader, 1024)

   	if err != nil {
   		return false, nil
	   	panic(err)
	}


	var publickey *rsa.PublicKey
	publickey = &privatekey.PublicKey

	// output to files
	privatekeyout, err := os.Create("/path/to/files/.private_key_mdfs")
	if err != nil {
		return false, nil
		panic(err)
	}
	encoder := gob.NewEncoder(privatekeyout)
	encoder.Encode(privatekey)
	privatekeyout.Close()

	publickeyout, err := os.Create("/path/to/files/.public_key_mdfs")
	if err != nil {
		return false, nil
		panic(err)
	}
	encoder = gob.NewEncoder(publickeyout)
	encoder.Encode(publickey)
	publickeyout.Close()

	return true, nil
}

func GenSymmetricKey() (key []byte, err error) {

	// create a byte array 32 bytes long
	key = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	return
}

func EncryptFile(filepath string, destination string) (err error) {

	var plaintext []byte
	var block cipher.Block

	if plaintext, err = ioutil.ReadFile(filepath); err != nil {
		panic(err)
	}

	ciphertext := make([]byte, 32+aes.BlockSize+len(string(plaintext)))
	//ciphertext := make([]byte, len(string(plaintext)))

	// PREPEND of user tokens will happen here, for now we will just
	// leave the key unencrypted

	// create key
	key := ciphertext[:32]
	if _, err = io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}

	// create initialization vector
	iv := ciphertext[32 : 32+aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	// create the cipher block from the key
	if block, err = aes.NewCipher(key); err != nil {
		panic(err)
	}

	// init an encryption stream
	encrypter := cipher.NewCTR(block, iv)

	encrypter.XORKeyStream(ciphertext[32+aes.BlockSize:], plaintext)
	//encrypter.XORKeyStream(ciphertext, plaintext)

	ioutil.WriteFile(destination, ciphertext, 0777)

	return
}

func CreateUserToken(block cipher.Block, iv []byte, privKey []string) {

}

func ProtectFile(filepath string) {

}
