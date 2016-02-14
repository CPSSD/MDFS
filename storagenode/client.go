package main

import (
	"MDFS/utils"
	"bufio"
	"net"
)

func main() {

	// doesn't get configuration from file
	// it will get it from metadata service
	protocol := "tcp"
	socket := "127.0.0.1:8081"

	// connect to this socket
	// there should probably be error checking here
	conn, _ := net.Dial(protocol, socket)

	// create a write buffer for the tcp connection
	w := bufio.NewWriter(conn)

	// create byte to hold handler code
	var code uint8
	code = 1

	// send code to the server
	err := w.WriteByte(code)
	if err != nil {
		panic(err)
	}

	// send file to server
	// hardcoded for testing purposes
	filepath := "/path/to/files/test"
	utils.SendFile(conn, w, filepath)
}
