package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/CPSSD/MDFS/utils"
	"net"
)

func main() {

	// command line flags
	var request string
	var send string

	flag.StringVar(&request, "request", "hash", "hash of file you are requesting")
	flag.StringVar(&send, "send", "filepath", "filepath to file you wish to send")

	flag.Parse()

	// doesn't get configuration from file
	// it will get it from metadata service
	protocol := "tcp"
	socket := "localhost:8081"

	// connect to this socket
	// there should probably be error checking here
	conn, _ := net.Dial(protocol, socket)
	defer conn.Close()

	// create a read and write buffer for the tcp connection
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	// create byte to hold handler code
	var sendcode uint8

	if request != "hash" {
		sendcode = 1
	} else if send != "filepath" {
		sendcode = 2
	} else {
		sendcode = 0
	}

	// send code to the server
	err := w.WriteByte(sendcode)
	if err != nil {
		panic(err)
	}

	switch sendcode {
	case 1: // requesting a file from storage node

		fmt.Println("Requesting: " + request)

		// get []byte representation of hash
		hash, err := hex.DecodeString(request)
		if err != nil {
			panic(err)
		}

		w.Write(hash)
		if err = w.Flush(); err != nil {
			panic(err)
		}

		// get handler code
		handlecode, _ := r.ReadByte()
		switch handlecode {
		case 3: // file available
			output := "./output"
			utils.ReceiveFile(conn, r, output)
		}

	case 2: // send file to server
		fmt.Println("Sending: " + send)
		utils.SendFile(conn, w, send)

	default:
		fmt.Println("Unrecognised code")
	}
}
