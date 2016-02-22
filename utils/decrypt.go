package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"io/ioutil"
)

func DecryptFile(filepath string, destination string) (err error) {

	var ciphertext []byte
	var block cipher.Block

	// Open the ciphertext to read
	if ciphertext, err = ioutil.ReadFile(filepath); err != nil {
		panic(err)
	}

	// REMOVAL of user tokens will happen here, for now we will just
	// assume the key is unencrypted

	// Get key from the ciphertext
	key := ciphertext[:32]

	// Create the AES cipher block from the key
	if block, err = aes.NewCipher(key); err != nil {
		panic(err)
	}

	// Init a GCM decrypter from the AES cipher
	decrypter, err := cipher.NewGCM(block)

	// Get the nonce from the ciphertext
	nonce := ciphertext[32 : 32+decrypter.NonceSize()]

	// Decrypt and authenticate the message to plaintext.
	// First nil arg is the destination, however the plaintext is
	// returned so we will store it in a byte array
	plaintext, err := decrypter.Open(nil, nonce, ciphertext[32+decrypter.NonceSize():], nil)
	if err != nil {
		panic(err)
	}

	// Write plaintext to destination file with permissions 0777
	ioutil.WriteFile(destination, plaintext, 0777)

	return
}

func ExtractToken(user User, tokens []byte) (token []byte, err error) {

	return
}
