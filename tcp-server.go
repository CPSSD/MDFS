package main

import (
	"MDFS/config"
	"MDFS/utils"
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
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
		go handleRequest(conn, conf.Path)
	}
}

// handles incoming requests
func handleRequest(conn net.Conn, path string) {

	// will listen for message to process ending in newline (\n)
	hash, _ := bufio.NewReader(conn).ReadString('\n')
	hash = strings.TrimSpace(string(hash))

	// output messsage received
	fmt.Println("Hash Received:", hash)

	// check if received hash value exists
	var message string
	if utils.CheckForHash(path, hash) {
		message = "file exists on storage node"
	} else {
		message = "file does not exist on storage node"
	}

	// send new string back to client
	conn.Write([]byte(message + "\n"))

	// close connection
	conn.Close()
}
