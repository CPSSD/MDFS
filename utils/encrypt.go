package utils

import (

    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    // "os"
    "io"
    // "fmt"
)

func GenCipherBlock() (block cipher.Block, err error)  {

    // create a byte array 32 bytes long
    data := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, data); err != nil {
        panic(err)
    }

    // get a cipher block from the key
    block, err = aes.NewCipher(data)
    if err != nil {
        panic(err)
    }
    return
}

func GenCipherTextAndKey(str string) (encrypted []byte, block cipher.Block, iv []byte) {
    
    // convert the string input to byte array
    // NEEDS TO BE CHANGED, SHOULD ACCEPT A BYTE ARRAY
    plaintext := []byte(str)
    
    // get the block required for encryption and decryption
    block, err := GenCipherBlock()
    if err != nil {
        panic(err)
    }

    // get the right block size for the plaintext
    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv = ciphertext[:aes.BlockSize]

    // read a random value for IV
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        panic(err)
    }

    encrypter := cipher.NewCFBEncrypter(block, iv)

    encrypted = make([]byte, len(str))
    encrypter.XORKeyStream(encrypted, plaintext)


    return

}

func CreateUserToken(block cipher.Block, privKey []string) {
    // should use the symmetric key and private 
    // key of all users to generate tokens for 

}

func ProtectFile() {

}

