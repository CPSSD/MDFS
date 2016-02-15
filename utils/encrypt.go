package utils

import (

    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
//    "os"
    "io"
    "fmt"
)

func GenSymmetricKey() (block cipher.Block, err error)  {

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

func GenCipherText(str string) {
    
    plaintext := []byte(str)
    
    block, err := GenSymmetricKey()
    if err != nil {
        panic(err)
    }
    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        panic(err)
    }

    encrypter := cipher.NewCFBEncrypter(block, iv)

    encrypted := make([]byte, len(str))
    encrypter.XORKeyStream(encrypted, plaintext)

    fmt.Printf("%s encrypted to %v with iv of %v\n", str, encrypted, iv)


}

func CreateUserToken(block cipher.Block, privKey []string) {
    // should use the symmetric key and private 
    // key of all users to generate tokens for 

}

func ProtectFile() {

}

