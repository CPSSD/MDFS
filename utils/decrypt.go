package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io/ioutil"
)

var (
	ErrDecryption	= errors.New("Error in decryption\n")
	ErrNoToken		= errors.New("No token matching your uuid\n")
)

func DecryptFile(filepath string, destination string, user User) (err error) {

	var ciphertext []byte
	var block cipher.Block

	// Open the ciphertext to read
	if ciphertext, err = ioutil.ReadFile(filepath); err != nil {
		return err
	}

	// Find out the size of the set of tokens
	bufTsize := ciphertext[:8]
	tokens_size, n := binary.Uvarint(bufTsize)
	if n <= 0 {
		return ErrDecryption
	}

	tokens := ciphertext[8 : 8+int(tokens_size)]

	// Convert the uuid to a byte array 8B or 64b long
	bufuuid := make([]byte, 8)
	_ = binary.PutUvarint(bufuuid, user.Uuid)

	// Extract the symmetric key from the set of tokens for this user
	symkey, err := extractKeyFromToken(bufuuid, user.Privkey, tokens)
	if err != nil {
		return err
	}

	// Create the AES cipher block from the key
	if block, err = aes.NewCipher(symkey); err != nil {
		return err
	}

	// Init a GCM decrypter from the AES cipher
	decrypter, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Get the nonce from the ciphertext
	nonce := ciphertext[8+int(tokens_size) : 8+int(tokens_size)+decrypter.NonceSize()]

	// Decrypt and authenticate the message to plaintext.
	// First nil arg is the destination, however the plaintext is
	// returned so we will store it in a byte array
	plaintext, err := decrypter.Open(nil, nonce, ciphertext[8+int(tokens_size)+decrypter.NonceSize():], nil)
	if err != nil {
		return err
	}

	// Write plaintext to destination file with permissions 0777
	ioutil.WriteFile(destination, plaintext, 0777)

	return nil
}

func extractKeyFromToken(uuid []byte, privatekey *rsa.PrivateKey, tokens []byte) (symkey []byte, err error) {

	for i := 0; i < (len(tokens)); i += 136 {
		token := tokens[i : i+136]

		if bytes.Equal(token[:8], uuid) {

			hash := sha256.New()
			symkey, err = rsa.DecryptOAEP(hash, rand.Reader, privatekey, token[8:], uuid)
			return
		}

	}
	return nil, ErrNoToken
}
