package server

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
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
	st.conf = config.ParseConfiguration(utils.GetUserHome() + "/.stnode/stnode_conf.json")
}

func (md *MDService) parseConfig() {
	md.conf = config.ParseConfiguration(utils.GetUserHome() + "/.mdservice/.mdservice_conf.json")
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
	case 1: // ls
		fmt.Println("In ls")

		lenArgs, _ := r.ReadByte()
		fmt.Printf("lenArgs = %v\n", lenArgs)

		msg := ""

		if lenArgs == 1 {
			files, err := ioutil.ReadDir(md.getPath())
			if err != nil {
				panic(err)
			}

			for _, file := range files {
				msg = msg + "," + file.Name()
			}
		}

		for i := 1; i < int(lenArgs); i++ {
			fmt.Printf("  in loop at pos %d ready to read\n", i)

			path, _ := r.ReadString('\n')
			path = strings.TrimSuffix(path, "\n")

			msg = msg + path + ":/"

			fmt.Printf("  in loop read in path: %s", path)

			files, err := ioutil.ReadDir(md.getPath() + path)
			if err != nil {
				panic(err)
			}

			for _, file := range files {
				msg = msg + "," + file.Name()
			}
		}

		w.WriteString(msg + ", ")
		w.Flush()

		fmt.Println("Fin ls")

	case 2: // mkdir
		fmt.Println("In mkdir")

		lenArgs, _ := r.ReadByte()
		fmt.Printf("lenArgs = %v\n", lenArgs)

		for i := 1; i < int(lenArgs); i++ {
			fmt.Printf("  in loop at pos %d ready to read\n", i)
			path, _ := r.ReadString('\n')
			fmt.Printf("  in loop read in path: %s", path)
			os.MkdirAll(md.getPath()+strings.TrimSpace(path), 0777)
		}
		fmt.Println("Fin mkdir")

	case 3: // rmdir
		fmt.Println("In rmdir")

		lenArgs, _ := r.ReadByte()
		fmt.Printf("lenArgs = %v\n", lenArgs)

		for i := 1; i < int(lenArgs); i++ {
			fmt.Printf("  in loop at pos %d ready to read\n", i)
			path, _ := r.ReadString('\n')
			fmt.Printf("  in loop read in path: %s", path)
			os.Remove(md.getPath() + "./" + strings.TrimSpace(path))
		}
		fmt.Println("Fin mkdir")

	case 4: // cd
	case 5: // send
	case 6: // request
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
	fmt.Println("Ready to read code")
	code, err := r.ReadByte()
	for code != 0 {

		fmt.Printf("Read code: %v\n", code)
		in.handleCode(code, conn, r, w)

		code, err = r.ReadByte()

	}

	fmt.Printf("Connection close with code of %v and err of: %v\n", code, err)
}
