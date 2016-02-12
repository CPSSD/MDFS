package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {

	// doesn't get configuration from file
	// it will get it from metadata service
	protocol := "tcp"
	socket := "127.0.0.1:8081"

	// connect to this socket
	conn, _ := net.Dial(protocol, socket)

	// read in input from stdin
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("File hash to request: ")
	text, _ := reader.ReadString('\n')

	// send to socket
	fmt.Fprint(conn, text+"\n")

	// listen for reply
	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("Message received from server: " + message)

	// close the connection
	conn.Close()
}
