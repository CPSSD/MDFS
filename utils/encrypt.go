package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var (
	// Custom error messages
	ErrNoPublKey    = errors.New("Privatekey exists, but no publickey\n")
	ErrNoPrivKey    = errors.New("Publickey exists, but no privatekey\n")
	ErrNoKeyPair    = errors.New("No key-pair exists\n")
	ErrKeyPairExist = errors.New("A user key-pair already exists\n")

	ErrInvalidArgs	= errors.New("Invalid arguments to function\n")

	ErrEncryption	= errors.New("Error in encryption\n")
)

func KeysExist() (success bool, err error) {
	// return true if the keys exist locally
	// return false if only one or no keys exist

	// If file exists, os.Stat will return data and err will be nil
	// See if private exists
	if _, err := os.Stat("/path/to/files/.private_key_mdfs"); err == nil {

		// See if public exists
		if _, err := os.Stat("/path/to/files/.public_key_mdfs"); err == nil {

			return true, ErrKeyPairExist
		}

		// Error as defined above
		return false, ErrNoPublKey
	}
	if _, err := os.Stat("/path/to/files/.public_key_mdfs"); err == nil {

		// Error as defined above
		return false, ErrNoPrivKey
	}

	return false, ErrNoKeyPair
}

func GenUserKeys() (success bool, err error) {

	// Generate a user's public and private key.

	// Make sure they do not exist already.
	if success, err := KeysExist(); err != ErrNoKeyPair {
		if err == ErrKeyPairExist {
			fmt.Printf("NOTE: \tDid not generate new keys because:\n\t%v\n", err)
		}
		return success, err
	}

	// if no keys exist, continue

	// Generate a new RSA private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return false, err
	}

	// Get the public RSA key from the private one above
	var publickey *rsa.PublicKey
	publickey = &privatekey.PublicKey

	// Create output file for the private key
	privatekeyout, err := os.Create("/path/to/files/.private_key_mdfs")
	if err != nil {
		return false, err
	}

	// Create a gob encoder for the private key file
	encoder := gob.NewEncoder(privatekeyout)

	// Encode private key to the gob encoder's stream (the file)
	encoder.Encode(privatekey)
	privatekeyout.Close()

	// Same process for outputting the public key to disk
	publickeyout, err := os.Create("/path/to/files/.public_key_mdfs")
	if err != nil {
		return false, err
	}

	encoder = gob.NewEncoder(publickeyout)
	encoder.Encode(publickey)
	publickeyout.Close()

	return true, err
}

func genSymmetricKey() (key []byte, err error) {

	// Not currently used
	// Create a byte array 32 bytes long
	key = make([]byte, 32)

	// Fill the array with crypto-secure random data
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	return
}

func EncryptFile(filepath string, destination string, users ...User) (err error) {

	// Current structure of the final ciphertext:
	// [ num of user tokens (8B) | ... user token(s) ... | nonce (12B) | ciphertext (variable length) ]

	var plaintext []byte
	var block cipher.Block

	// Open the file to be encrypted (the plaintext)
	if plaintext, err = ioutil.ReadFile(filepath); err != nil {
		return err
	}

	// Generate a symmetric AES-256 key
	symkey, err := genSymmetricKey()
	if err != nil {
		return err
	}

	// Generate token(s) for the key
	tokens, err := prepTokens(symkey, users...)
	if err != nil {
		return err
	}

	tokens_size := make([]byte, 8)
	_ = binary.PutUvarint(tokens_size, uint64(len(tokens)))

	ciphertext := append(tokens_size, tokens...)

	// Create the AES cipher block from the key
	if block, err = aes.NewCipher(symkey); err != nil {
		return err
	}

	// Init a GCM (Galois/Counter Mode) encrypter from the AES cipher.
	encrypter, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Create a nonce (random data used in the encryption process).
	// The nonce used in encryption must be the same used in the
	// decryption process. Append it to ciphertext
	nonce := make([]byte, encrypter.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext = append(ciphertext, nonce...)

	// Seal appends the encrypted authenticated plaintext to ciphertext.
	// The nil value is optional data which is not being used currently.
	ciphertext = encrypter.Seal(ciphertext, nonce, plaintext, nil)

	// Write ciphertext to destination with permissions 0777
	ioutil.WriteFile(destination, ciphertext, 0777)

	return nil
}

type User struct {
	Uuid    uint64
	Pubkey  *rsa.PublicKey
	Privkey *rsa.PrivateKey
}

func prepTokens(symkey []byte, users ...User) (tokens []byte, err error) {

	if users == nil || symkey == nil {
		return nil, ErrInvalidArgs
	}

	// Loop through all of the users that are entered in args
	for i := 0; i < len(users); i++ {

		// Convert the uuid to a byte array 8B or 64b long
		buf := make([]byte, 8)
		_ = binary.PutUvarint(buf, users[i].Uuid)

		// Create a single token
		token, err := createUserToken(buf, users[i].Pubkey, symkey)
		if err != nil {
			return nil, err
		}

		// Append the token to the list of tokens
		tokens = append(tokens, token...)
	}

	return
}

func createUserToken(uuid []byte, publickey *rsa.PublicKey, symkey []byte) (token []byte, err error) {

	// Get a new sha256 hash for randomness
	hash := sha256.New()

	// Pass in hash function, random reader for entropy, user's public key,
	// the symkey (or data) to be encrypted, and the unique user id as a
	// label used in verification
	encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, publickey, symkey, uuid)
	if err != nil {
		return nil, err
	}
	token = append(uuid, encrypted...)

	return
}