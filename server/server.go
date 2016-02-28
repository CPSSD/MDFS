package server

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"io"
	"net"
	"os"
)

type Server struct {
	conf config.Configuration
}

type StorageNode struct {
	Server // anonymous field of type Server
}

type MDService struct {
	Server // anonymous field of type Server
}

// the Server interface
type TCPServer interface {
	parseConfig()
	getPath() string
	getProtocol() string
	getPort() string
	getHost() string
	handleCode(code uint8, conn net.Conn, r *bufio.Reader, w *bufio.Writer)
}

// Server methods
func (s Server) getPath() string {
	return s.conf.Path
}

func (s Server) getProtocol() string {
	return s.conf.Protocol
}

func (s Server) getPort() string {
	return s.conf.Port
}

func (s Server) getHost() string {
	return s.conf.Host
}

func (st *StorageNode) parseConfig() {
	st.conf = config.ParseConfiguration(utils.GetUserHome() + "/path/to/files/config/stnode_conf.json")
}

func (md *MDService) parseConfig() {
	md.conf = config.ParseConfiguration(utils.GetUserHome() + "/path/to/files/config/mdservice_conf.json")
}

// checks request code and calls corresponding function
func (st StorageNode) handleCode(code uint8, conn net.Conn, r *bufio.Reader, w *bufio.Writer) {

	switch code {
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
		fp := st.getPath() + hash
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
		output := st.getPath() + "output"
		utils.ReceiveFile(conn, r, output)
		hash, err := utils.ComputeMd5(output)
		if err != nil {
			panic(err)
		}
		checksum := hex.EncodeToString(hash)
		fmt.Println("md5 checksum of file is: " + checksum)
		os.Rename(output, st.getPath()+checksum)
	}
}

func (md MDService) handleCode(code uint8, conn net.Conn, r *bufio.Reader, w *bufio.Writer) {

	switch code {
	default:
		return
	}
}

// package functions
func Start(in TCPServer) {

	// get server configuration from json file
	in.parseConfig()
	fmt.Println("Launching Server...")

	protocol := in.getProtocol()
	host := in.getHost()
	port := in.getPort()

	// listen on specified interface & port
	ln, err := net.Listen(protocol, host+":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		panic(err)
	}
	defer ln.Close() // close the listener when the function exits

	fmt.Println("Listening on " + in.getPort())

	// run loop forever (or until ctrl-c)
	for {

		// accept connection on port
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accpting:", err.Error())
			panic(err)
		}

		// handle connection in new goroutine
		go handleRequest(conn, in)
	}
}

func handleRequest(conn net.Conn, in TCPServer) {

	defer conn.Close()

	// create read and write buffer for tcp connection
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	// var code uint8
	code, _ := r.ReadByte()

	in.handleCode(code, conn, r, w)
}
