#Utils Package
This package contains all the utility functions we have created. These include encryption/decryption functions, send/receive files, generating file hashes and functions pertaining to persistence.

##hash.go
This file contains all the necessary functions for dealing with file hashes.

####ComputeMd5() function
This function takes a string parameter to represent a filepath and returns a byte array and an error. The byte array should contain the byte representation of the file's md5 checksum. It uses the crypto/md5 package.

```Go
// returns the the files md5 checksum in a byte array
// the checksum should only be used if the error returned is nil
func ComputeMd5(filepath string) ([]byte, error) {

    // byte array to hold md5 sum
    var result []byte
    file, err := os.Open(filepath)
    if err != nil {
        return result, err
    }
    defer file.Close()

    hash := md5.New() // type hash.Hash
    // copy the contents of the file into hash
    if _, err := io.Copy(hash, file); err != nil {
        return result, err
    }

    return hash.Sum(result), nil
}
```

####CheckForHash()
This function takes a path to a directory and a hash in the form of strings and returns a bool to indicate whether a file named after the hash is present in the specified directory. This function is used by the storage node as the storage node names the files it stores after their hash.

```Go
// used by storage node to find files
func CheckForHash(path, hash string) bool {

    if _, err := os.Stat(path + hash); err == nil {
        return true
    } else {
        return false
    }
}
```

####ReadHashAsString() function
This function takes a pointer to a bufio.Reader and returns a hex representation of hash in string format. A 16 byte buffer is created to hold the hash and the reader fills it with the bytes read. The buffer is then encoded to a hex string using the hex.EncodeToString() function.

```Go
// reads bytes from a bufio reader and returns the stirng representation
func ReadHashAsString(r *bufio.Reader) string {

    // make a buffer to hold hash
    buf := make([]byte, 16)
    _, err := r.Read(buf)
    if err != nil && err != io.EOF {
        panic(err)
    }

    hash := hex.EncodeToString(buf)
    return hash
}
```

####WriteHash() function
This function takes a pointer to a bufio.Reader and a hash in the form of a byte array and returns an error to indicate success/failure.

```Go
// writes a byte array, representing a hash, to a bufio writer
func WriteHash(w *bufio.Writer, hash []byte) error {

    w.Write(hash)
    return w.Flush()
}
```

##file_transfer.go
This file contains all the necessary functions for sending/receiving files over a network connection.