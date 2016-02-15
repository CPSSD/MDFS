package main

import (
	"net"
	"MDFS/utils"
	"fmt"
)

func main() {
	
	// encryption of a string
	str := "Hello World"
	encrypted, block, iv := utils.GenCipherTextAndKey(str)
	
    fmt.Printf("%s encrypted to %v with iv of %v and block of %v\n", str, encrypted, iv, block)
	// doesn't get configuration from file
	// it will get it from metadata service
	protocol := "tcp"
	socket := "127.0.0.1:8081"

	// connect to this socket
	// there should probably be error checking here
	conn, _ := net.Dial(protocol, socket)

	// send file to server
	// hardcoded for testing purposes
	filepath := "/path/to/files/input.jpg"
	utils.SendFile(conn, filepath)
}
