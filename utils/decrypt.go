package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "io/ioutil"
    "fmt"
)

func DecryptFile(filepath string, destination string) (err error) {

    var ciphertext []byte
    var block cipher.Block

    if ciphertext, err = ioutil.ReadFile(filepath); err != nil {
        panic(err)
    }

    //plaintext := make([]byte, 32+aes.BlockSize+len(string(ciphertext)))

    // PREPEND of user tokens will happen here, for now we will just
    // leave the key unencrypted

    // get key
    key := ciphertext[:32]
    fmt.Printf("\n\nThe key read is: %v \n\n", key)

    // create the cipher block from the key
    if block, err = aes.NewCipher(key); err != nil {
        panic(err)
    }

    // init a decryption stream
    // decrypter := cipher.NewCTR(block, iv)
    decrypter, err := cipher.NewGCM(block)
/*
    // get initialization vector
    
    iv := ciphertext[32:32+aes.BlockSize]
    if len(ciphertext) < aes.BlockSize {
        panic(err)
        return
    }
*/
    // get the nonce from the ciphertext
    nonce := ciphertext[32:32+decrypter.NonceSize()]
    fmt.Printf("\n\nThe nonce read is: %v \n\n", nonce)

    // remove the key and iv from the ciphertext
    // // // // ciphertext = ciphertext[32+decrypter.NonceSize():]


    // decryption can be done in place
    // decrypter.XORKeyStream(ciphertext, ciphertext)

    // for clarity's sake
    plaintext, err := decrypter.Open(nil, nonce, ciphertext[32+decrypter.NonceSize():], nil)
    if err != nil {
        panic(err)
    }

    ioutil.WriteFile(destination, plaintext, 0777)

    return
}