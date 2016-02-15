package utils

import (
    "crypto/cipher"
)

func Decrypt(encrypted []byte, block cipher.Block, iv []byte) ([]byte) {

    // create a stream to decrypt according to block and IV
    decrypter := cipher.NewCFBDecrypter(block, iv)

    // decrypt encrypted to encrypted. Can work in place because args
    // are the same.
    decrypter.XORKeyStream(encrypted, encrypted)

    // although called encrypted, it is the decrypted message
    return encrypted
}