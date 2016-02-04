package main

import (
	"net"
	"fmt"
	"bufio"
	"os"
)

func main() {

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:8081")

	for {

		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')

		// send to socket
		fmt.Fprint(conn, text + "\n")

		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message received from server: " + message)
	}
}