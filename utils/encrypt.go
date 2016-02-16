package utils

import (

    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "os"
    "io"
    "fmt"
    "bufio"
    "encoding/binary"
)

func GenSymmetricKey() (key []byte, err error)  {

    // create a byte array 32 bytes long
    key := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, key); err != nil {
        panic(err)
    }
    return
}

func GenCipherText(filepath string)  {
    
    // open file to encrypt
    fi, err := os.Open(filepath)
    if err != nil {
        panic(err)
    }
    defer fi.Close()

    // location of new encrypted file
    fo, err := os.Create(filepath+".enc")
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
    
    // get the block required for encryption and decryption
    block, iv, err := GenCipherBlockAndIV()
    if err != nil {
        panic(err)
    }

    // HERE THE CIPHER BLOCK AND INIT VECTOR
    // SHOULD BE ENCRYPTED AS USER TOKEN PAIRS
    // For now we will simply prepend them

    n, err := w.Write(iv)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%d bytes written for iv\n", n)

    bo := []byte(fmt.Sprintf("%v", block))
    bl := uint16((len(bo)))

    err = binary.Write(w, binary.LittleEndian, bl)
    if err != nil {
        panic(err)
    }
    fmt.Printf("some bits written for bl %d\n", bl)

    n, err = w.Write(bo)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%d bytes written for bo of len%d\n", n, len(bo))


    // init an encryption stream
    encrypter := cipher.NewCFBEncrypter(block, iv)

    //encrypted = make([]byte, len(str))


    for {

        // read a chunk
        n, err := r.Read(buf)
        if err != nil && err != io.EOF {
            panic(err)
        }
        if n==0 {
            break
        }

        // encrypt the chunk
        encrypter.XORKeyStream(buf, buf)

        // write the chunk to the new file
        if _, err := w.Write(buf[:n]); err != nil {
            panic(err)
        }
    }

    if err = w.Flush(); err != nil {
        panic(err)
    }
}

func CreateUserToken(block cipher.Block, iv []byte, privKey []string) {
    // should use the symmetric key and private 
    // key of all users to generate tokens for 

}

func ProtectFile(filepath string) {


}