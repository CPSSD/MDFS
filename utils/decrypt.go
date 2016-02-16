package utils

import (
    "crypto/cipher"
    "os"
    "io"
    "fmt"
    "bufio"
    "encoding/binary"
)

func Decrypt(filepath string) {
    
    // open file to decrypt
    fi, err := os.Open(filepath)
    if err != nil {
        panic(err)
    }
    defer fi.Close()

    // location of new decrypted file
    fo, err := os.Create(filepath+".dec")
    if err != nil {
        panic(err)
    }
    defer fo.Close()

    r := bufio.NewReader(fi)

    w := bufio.NewWriter(fo)

    buf := make([]byte, 1024)

    // convert the string input to byte array
    // NEEDS TO BE CHANGED, SHOULD ACCEPT A BYTE ARRAY
    //plaintext := []byte(str)
    
    // HERE THE CIPHER BLOCK AND INIT VECTOR
    // SHOULD BE ENCRYPTED AS USER TOKEN PAIRS
    // For now we will simply prepend them

    var blen uint16

    iv := make([]byte, 2)
    n, err := r.Read(iv)
    if err != nil {
        panic(err)
    }
    fmt.Printf("bytes read for iv %d\n", iv)    

    err = binary.Read(r, binary.LittleEndian, &blen)
    if err != nil {
        panic(err)
    }
    fmt.Printf("some bits read for bl %d\n", blen)


    //bo := []byte(fmt.Sprintf("%v", block))
    bo := make([]byte, blen)

    for {
        n, err := r.Read(bo)
        if err != nil && err != io.EOF {
            panic(err)
        }
        if n==0 {
            break
        }
        fmt.Printf("%d bytes read for bo of len%d\n", n, len(bo))
    }

    block := cipher.Block(bo)

    // create a stream to decrypt according to block and IV
    decrypter := cipher.NewCFBDecrypter(block, iv)

    for {

        // read a chunk
        n, err := r.Read(buf)
        if err != nil && err != io.EOF {
            panic(err)
        }
        if n==0 {
            break
        }

        // decrypt encrypted to encrypted. Can work in place because args
        // are the same.
        decrypter.XORKeyStream(encrypted, encrypted)

        // write the chunk to the new file
        if _, err := w.Write(buf[:n]); err != nil {
            panic(err)
        }
    }

    if err = w.Flush(); err != nil {
        panic(err)
    }
}