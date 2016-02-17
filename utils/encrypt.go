package utils

import (

    "fmt"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "io"
    "io/ioutil"
)

func GenSymmetricKey() (key []byte, err error)  {

    // create a byte array 32 bytes long
    key = make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, key); err != nil {
        panic(err)
    }
    return
}

func EncryptFile(filepath string, destination string) (iv []byte, key []byte, err error) {

    var plaintext []byte
    var block cipher.Block

    if plaintext, err = ioutil.ReadFile(filepath); err != nil {
        panic(err)
    }

    //ciphertext := make([]byte, 32+aes.BlockSize+len(string(plaintext)))
    ciphertext := make([]byte, len(string(plaintext)))

    // PREPEND of user tokens will happen here, for now we will just
    // leave the key unencrypted

    // create key
/*
    key := ciphertext[:32]
    if _, err = io.ReadFull(rand.Reader, key); err != nil {
        panic(err)
    }
*/
    key, err = GenSymmetricKey()

    // create initialization vector
/*
    iv := ciphertext[32:32+aes.BlockSize]
    if _, err = io.ReadFull(rand.Reader, iv); err != nil {
        panic(err)
    }
*/
    iv = make([]byte, aes.BlockSize)
    if _, err = io.ReadFull(rand.Reader, iv); err != nil {
        panic(err)
    }

    key = []byte("longer means more possible keys ")
    iv = []byte("longer means mor")

    // create the cipher block from the key
    if block, err = aes.NewCipher(key); err != nil {
        panic(err)
    }
    
    fmt.Printf("%d", aes.BlockSize)

    // init an encryption stream
    encrypter := cipher.NewCTR(block, iv)

    //encrypter.XORKeyStream(ciphertext[32+aes.BlockSize:], plaintext)
    encrypter.XORKeyStream(ciphertext, plaintext)

    ioutil.WriteFile(destination, ciphertext, 0777)

    return
}

func CreateUserToken(block cipher.Block, iv []byte, privKey []string) {
     
}

func ProtectFile(filepath string) {

}