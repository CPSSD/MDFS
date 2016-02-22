package main

import (
    "fmt"
    //"io"
    "os"
    "crypto/rsa"
    "encoding/gob"
)

func main() {

    var prk rsa.PrivateKey
    var puk rsa.PublicKey

    prf, err := os.Open("/path/to/files/.private_key_mdfs")
    if err != nil {
        panic(err)
    }
    puf, err := os.Open("/path/to/files/.public_key_mdfs")
    if err != nil {
        panic(err)
    }

    decoder := gob.NewDecoder(prf)
    decoder.Decode(&prk)
    prf.Close()
    fmt.Printf("Opened private key file: \n%v\n", prk)

    decoder = gob.NewDecoder(puf)
    decoder.Decode(&puk)
    puf.Close()
    fmt.Printf("Opened public key file: \n%v\n", puk)
}