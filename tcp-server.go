package main

import (
	"MDFS/config"
	"MDFS/utils"
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {

	fmt.Println("Launching server...")

	// get configuration settings from config file
	// TODO relative vs abosolute file path
	// requires absolute fp when installed
	conf := config.ParseConfiguration("config/tcp-server-conf.json")

	// listen on specified port
	ln, _ := net.Listen(conf.Protocol, conf.Port)

	// accept connection on port
	conn, _ := ln.Accept()

	// run loop forever (or until ctrl-c)
	for {

		// will listen for message to process ending in newline (\n)
		hash, _ := bufio.NewReader(conn).ReadString('\n')
		hash = strings.TrimSpace(string(hash))

		// output messsage received
		fmt.Println("Hash Received:", hash)

		// check if received hash value exists
		var message string
		if utils.CheckForHash(conf.Path, hash) {
			message = "file exists on storage node"
		} else {
			message = "file does not exist on storage node"
		}

		// send new string back to client
		conn.Write([]byte(message + "\n"))
	}
}
