package main

import (
	"MDFS/config"
	"MDFS/utils"
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
)

// get configuration settings from config file
// TODO relative vs abosolute file path
// requires absolute fp when installed
var conf = config.ParseConfiguration("../config/serverconf.json")

func main() {

	fmt.Println("Launching server...")

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

	defer conn.Close()

	// create read and write buffer for tcp connection
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	// var code uint8
	handlecode, _ := r.ReadByte()

	// should there be a confirmation sent from server to client?

	switch handlecode {
	case 1: // client is requesting a file
		// make a buffer to hold hash
		buf := make([]byte, 16)
		_, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		hash := hex.EncodeToString(buf)
		fmt.Println("Hash received: " + hash)

		// check if file exists
		var sendcode uint8
		fp := conf.Path + hash
		if _, err := os.Stat(fp); err == nil {
			sendcode = 3                 // file available code
			err := w.WriteByte(sendcode) // let client know
			if err != nil {
				panic(err)
			}

			// send the file
			utils.SendFile(conn, w, fp)
		} else {
			sendcode = 4
			err := w.WriteByte(sendcode) // let client know
			if err != nil {
				panic(err)
			}
		}

	case 2: // receive file from client
		output := conf.Path + "output"
		utils.ReceiveFile(conn, r, output)
		hash, err := utils.ComputeMd5(output)
		if err != nil {
			panic(err)
		}
		checksum := hex.EncodeToString(hash)
		fmt.Println("md5 checksum of file is: " + checksum)
		os.Rename(output, conf.Path+checksum)
	}
}
