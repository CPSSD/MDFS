package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var (
	ErrNoPublKey    = errors.New("Privatekey exists, but no publickey.")
	ErrNoPrivKey    = errors.New("Publickey exists, but no privatekey.")
	ErrNoKeyPair    = errors.New("No key-pair exists.")
	ErrKeyPairExist = errors.New("A user key-pair already exists.")
)

func KeysExist() (success bool, err error) {
	// return true if the keys exist locally
	// return false if only one or no keys exist

	if _, err := os.Stat("/path/to/files/.private_key_mdfs"); err == nil {

		// /path/to/files/.private_key_mdfs exists
		if _, err := os.Stat("/path/to/files/.public_key_mdfs"); err == nil {

			return true, ErrKeyPairExist
		}
		return false, ErrNoPublKey
	}
	if _, err := os.Stat("/path/to/files/.public_key_mdfs"); err == nil {
		// path/to/whatever exists
		return false, ErrNoPrivKey
	}

	return false, ErrNoKeyPair
}

func GenUserKeys() (success bool, err error) {
	// Generate a user's public and private key.
	// This should only be called if they do not exist already.
	if success, err := KeysExist(); err != ErrNoKeyPair {
		if err == ErrKeyPairExist {
			fmt.Printf("NOTE: \tDid not generate new keys because:\n\t%v\n", err)
		}
		return success, err
	}

	// if no keys exist, continue

	privatekey, err := rsa.GenerateKey(rand.Reader, 1024)

	if err != nil {
		return false, err
	}

	var publickey *rsa.PublicKey
	publickey = &privatekey.PublicKey

	// output to files
	privatekeyout, err := os.Create("/path/to/files/.private_key_mdfs")
	if err != nil {
		return false, err
	}
	encoder := gob.NewEncoder(privatekeyout)
	encoder.Encode(privatekey)
	privatekeyout.Close()

	publickeyout, err := os.Create("/path/to/files/.public_key_mdfs")
	if err != nil {
		return false, err
	}
	encoder = gob.NewEncoder(publickeyout)
	encoder.Encode(publickey)
	publickeyout.Close()

	return true, err
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

	// 44 = 32 byte key for aes + 12 bytes for the nonce
	ciphertext := make([]byte, 44)
	// PREPEND of user tokens will happen here, for now we will just
	// leave the key unencrypted

	// create AES-256 key
	key := ciphertext[:32]
	if _, err = io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}

	// create the cipher block from the key
	if block, err = aes.NewCipher(key); err != nil {
		panic(err)
	}

	// init a GCM encrypter
	encrypter, err := cipher.NewGCM(block)

	// create a nonce
	nonce := ciphertext[32 : 32+encrypter.NonceSize()]
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	// Seal appends the encrypted authenticated message to ciphertext
	ciphertext = encrypter.Seal(ciphertext, nonce, plaintext, nil)

	// Write encrypted data to file
	ioutil.WriteFile(destination, ciphertext, 0777)

	return
}

func CreateUserToken(uid []byte, publickey *rsa.PublicKey, symkey []byte) (token []byte, err error) {
	hash := sha256.New()
	encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, publickey, symkey, uid)
	if err != nil {
		return nil, err
	}
	token = append(uid, encrypted...)

	return
}

func ProtectFile(filepath string) {

}
