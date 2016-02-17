package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "io/ioutil"
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

    // get initialization vector
    
    iv := ciphertext[32:32+aes.BlockSize]
    if len(ciphertext) < aes.BlockSize {
        panic(err)
        return
    }




    // create the cipher block from the key
    if block, err = aes.NewCipher(key); err != nil {
        panic(err)
    }
    
    // remove the key and iv from the ciphertext
    ciphertext = ciphertext[32+aes.BlockSize:]

    // init an encryption stream
    decrypter := cipher.NewCTR(block, iv)

    // decryption can be done in place
    decrypter.XORKeyStream(ciphertext, ciphertext)

    // for clarity's sake
    plaintext := ciphertext

    ioutil.WriteFile(destination, plaintext, 0777)

    return
}