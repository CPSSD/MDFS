package main

import (
	//"net"
	"MDFS/utils"
	//"fmt"
)

func main() {
	
	// encryption of a string
	//str := "Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; In id pellentesque eros. Proin ut vulputate magna. Pellentesque elementum sem eu nibh finibus, id sodales orci efficitur. Donec viverra semper diam a tristique. Aliquam ut augue vestibulum, cursus erat nec, lacinia magna. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed neque nisl, rhoncus nec velit id, ornare mollis augue. Praesent imperdiet ut massa vitae varius."
	//encrypted, block, iv := utils.GenCipherTextAndKey(str)
	source := "/path/to/files/input.txt"
	encryp := "/path/to/files/input.enc"
	result := "/path/to/files/result.txt"

	iv, key, _ := utils.EncryptFile(source, encryp)
	utils.DecryptFile(iv, key, encryp, result)

	source = "/path/to/files/image.jpg"
	encryp = "/path/to/files/image.enc"
	result = "/path/to/files/result.jpg"

	iv, key, _ = utils.EncryptFile(source, encryp)
	utils.DecryptFile(iv, key, encryp, result)
	
    //fmt.Printf("%s encrypted to %v with iv of %v and block of %v\n", str, encrypted, iv, block)

    //result := utils.Decrypt(encrypted, block, iv)
    //plain := string(result)
    //fmt.Printf("%v decrypted to %s\n", encrypted, plain)
	

/*
	// doesn't get configuration from file
	// it will get it from metadata service
	protocol := "tcp"
	socket := "127.0.0.1:8081"

	// connect to this socket
	// there should probably be error checking here
	conn, _ := net.Dial(protocol, socket)

	// send file to server
	// hardcoded for testing purposes
	utils.SendFile(conn, filepath)
*/
}