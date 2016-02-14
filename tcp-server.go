package main

import (
	"MDFS/config"
	"MDFS/utils"
	"fmt"
	"net"
	"os"
)

func main() {

	fmt.Println("Launching server...")

	// get configuration settings from config file
	// TODO relative vs abosolute file path
	// requires absolute fp when installed
	conf := config.ParseConfiguration("config/tcp-server-conf.json")

	// listen on specified port
	ln, err := net.Listen(conf.Protocol, conf.Port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// close the listener when the application closes
	defer ln.Close()
	fmt.Println("Listening on " + conf.Port)

	// run loop forever (or until ctrl-c)
	for {

		// accept connection on port
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accpting:", err.Error())
			os.Exit(1)
		}

		// handle connection in new goroutine
		go handleRequest(conn)
	}
}

// checks request code and calls corresponding function
func handleRequest(conn net.Conn) {

	// will eventually check codes
	// only handles receiving files right now
	output := "/path/to/files/output.jpg"
	utils.ReceiveFile(conn, output)
}