package server

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"github.com/boltdb/bolt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
)

type Server struct {
	conf config.Configuration
}

type StorageNode struct {
	Server // anonymous field of type Server
}

type MDService struct {
	userDB   bolt.DB
	stnodeDB bolt.DB
	Server   // anonymous field of type Server
}

// the Server interface
type TCPServer interface {
	parseConfig()
	setup()
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

func (st StorageNode) setup() {

}

func (md MDService) setup() {

	// init the boltdb if it is not existant already
	// one for users, one for stnodes
	fmt.Println("This is a metadata service, opening DB's")
	userDB, err := bolt.Open(md.getPath()+".userDB.db", 0777, nil)
	if err != nil {
		panic(err)
	}
	defer userDB.Close()

	stnodeDB, err := bolt.Open(md.getPath()+".stnodeDB.db", 0777, nil)
	if err != nil {
		panic(err)
	}
	defer stnodeDB.Close()
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

	// switch statement for commands
	// commands work but are not yet optimized, possible code duplication occurs,
	// although moving some code to a function to reduce duping may not be worth it
	// if it is only one or two lines.
	switch code {
	case 1: // ls
		fmt.Println("In ls")

		// get current dir
		// NOTE: here and in other locations, trimming whitespace may be more desirable
		currentDir, _ := r.ReadString('\n')
		currentDir = strings.TrimSuffix(currentDir, "\n")

		// get the length of arguments to the ls command
		lenArgs, _ := r.ReadByte()
		fmt.Printf("lenArgs = %v\n", lenArgs)

		// the message which will be returned to the user, each entry will be
		// comma separated for the client to interpret
		msg := ""

		// if only the ls command was called
		if lenArgs == 1 {

			// md.getPath() == $HOME/.mdservice/
			// NOTE: currentDir should always start with a "/" and end in a normal char,
			// ie. not a "/". Following this convention avoids occurrences of double
			// slashes, missing slashes etc.

			// read the contents of the directory location to files
			// BUG-NOTE: from client, if you cd into a directory, and this directory is
			// subsequesntly deleted (by os user or otherwise) you will be unable to cd
			// or ls back up as your currentDir will evaluate to a dir that does not exist.
			// Possible fix is to do the conversion from relative to absolute path on the
			// client side (if possible, not sure if it is or not), perhaps failsafe and
			// return user to home ("/"), or maybe accept that deletions of dirs will
			// likely not occurr when demoing
			files, err := ioutil.ReadDir(md.getPath() + "files/" + currentDir)
			if err != nil {
				w.Flush()
			}

			// iterate over the files, and comma separate them while appending to msg
			for _, file := range files {
				msg = msg + file.Name() + ","
			}

		}

		// loop for dealing with one or more args
		for i := 1; i < int(lenArgs); i++ {

			fmt.Printf("  in loop at pos %d ready to read\n", i)

			// reading in this arg
			targetPath, _ := r.ReadString('\n')
			targetPath = strings.TrimSuffix(targetPath, "\n")

			fmt.Printf("  in loop read in targetPath: %s\n", (currentDir + "/" + targetPath))

			// added in a "/" as we do not know if the user tried to call "ls dir" or "ls /dir"
			// or "ls ./dir" etc. ReadDir() does not mind extra "/"s. "ls /dir" and "ls ./dir"
			// currently evaluate to the same thing as they are both prepended by currentDir
			// below, possibly viewed as a bug because it is not identical to UNIX ls.
			files, err := ioutil.ReadDir(md.getPath() + "files/" + currentDir + "/" + targetPath)
			if err != nil {

				// if it is not a directory, skip it and try the next arg
				continue

			} else {

				// allows multiple dirs to be ls'd and still know which is which
				//
				// ex. output:
				//
				// jim:/memes >> ls pepe nyan
				// pepe:
				// img1.jpg
				// img2.png
				//
				// nyan:
				// sound.wav
				//
				// jim:/memes >>

				msg = msg + targetPath + ":," // note comma to denote newline

				for _, file := range files {
					msg = msg + file.Name() + ","
				}

				// add an extra newline for spacing on client side
				msg = msg + ","
			}
		}

		// remove the last newline for graphical reasons
		msg = strings.TrimSuffix(msg, ",")

		// add the newline back in for some bizarre reaason? not sure why,
		// don't remove without testing effects on all types of ls command
		w.WriteString(msg + ", ")
		w.Flush()

		// print for terminal's sake
		fmt.Println("Fin ls")

	case 2: // mkdir
		fmt.Println("In mkdir")

		// get currentDir
		currentDir, _ := r.ReadString('\n')
		currentDir = strings.TrimSuffix(currentDir, "\n")

		// get lenArgs for mkdir
		lenArgs, _ := r.ReadByte()
		fmt.Printf("lenArgs = %v\n", lenArgs)

		// if no more than "mkdir" is sent, nothing will happen
		// if errors occur, the dir will just not be made
		for i := 1; i < int(lenArgs); i++ {

			fmt.Printf("  in loop at pos %d ready to read\n", i)

			// for each arg, get the target path
			targetPath, _ := r.ReadString('\n')

			// print the target for terminal's sake
			fmt.Printf("  in loop read in targetPath: %s", targetPath)

			// MkdirAll creates an entire file path if some dirs are missing
			os.MkdirAll(md.getPath()+"files/"+currentDir+"/"+strings.TrimSpace(targetPath), 0777)
		}

		// end of mkdir
		fmt.Println("Fin mkdir")

	case 3: // rmdir
		fmt.Println("In rmdir")

		// get currentDir
		currentDir, _ := r.ReadString('\n')
		currentDir = strings.TrimSuffix(currentDir, "\n")

		// get len args for rmdir
		lenArgs, _ := r.ReadByte()
		fmt.Printf("lenArgs = %v\n", lenArgs)

		// only does something if more than "rmdir" is called, as above
		for i := 1; i < int(lenArgs); i++ {

			fmt.Printf("  in loop at pos %d ready to read\n", i)

			targetPath, _ := r.ReadString('\n')
			fmt.Printf("  in loop read in targetPath: %s", targetPath)

			// this will only remove a dir that is empty, else it does nothing
			// BUG-NOTE: this command will also currently delete files (there is not
			// a different command to rmdir an rm in golang), so a check to make sure
			// the targetPath is a dir should take place (sample code for checking if
			// a path is a dir or a file is found in "cd" below).
			// NOTE: a nice to have would be a recursive remove similar to rm -rf,
			// but this is not needed
			os.Remove(md.getPath() + "files/" + currentDir + "/" + strings.TrimSpace(targetPath))
		}

		// end of rmdir
		fmt.Println("Fin rmdir")

	case 4: // cd
		fmt.Println("In cd")

		// get current dir
		currentDir, _ := r.ReadString('\n')
		currentDir = strings.TrimSuffix(currentDir, "\n")

		// print for terminal's sake
		fmt.Printf("currentDir = %s\n", currentDir)

		// get target path
		targetPath, _ := r.ReadString('\n')
		targetPath = strings.TrimSuffix(targetPath, "\n")

		// print for terminal's sake
		fmt.Printf("targetPath = %s\n", targetPath)

		// remove leading and trailing "/"s, not sure if this was necessary,
		// test before removal for all possible cd commands
		targetPath = strings.TrimPrefix(targetPath, "/")
		targetPath = strings.TrimSuffix(targetPath, "/")

		// check if the source dir exist
		src, err := os.Stat(md.getPath() + "files/" + currentDir + "/" + targetPath)
		if err != nil { // not a path ie. not a dir OR a file

			fmt.Println("Path is not a directory")

			// notify the client that it is not a dir with error code "1"
			w.WriteByte(1)
			w.Flush()

		} else { // is a path, but is it a dir or a file?

			// check if the source is indeed a directory or not
			if !src.IsDir() {

				fmt.Println("Path is not a directory")

				// notify the client that it is not a dir with error code "1"
				w.WriteByte(1)
				w.Flush()

			} else { // success!

				// notify success to client (no specific code, just not 1 or 0)
				w.WriteByte(2)
				w.Flush()

				// create a clean path that the user can display on the cmd line
				targetPath := path.Join(currentDir + "/" + targetPath)
				fmt.Printf("Path \"%s\" is a directory\n", targetPath)

				// send the new path back to the user
				w.WriteString(targetPath + "\n")
				w.Flush()
			}
		}

		// the below cases will entail the logging of a file in the mdservice, telling
		// the client which storage node to use, where to access files, sending public
		// keys, permissions, etc.

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

	// mdservice would initialise database here
	in.setup()

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

	// read in the handling code from the connected client
	code, err := r.ReadByte()
	// as long as there is no error in the code reading in..
	for code != 0 {

		// Print the code to terminal
		fmt.Printf("Read in code: %v\n", code)
		in.handleCode(code, conn, r, w)

		// wait to read the next code from the client
		code, err = r.ReadByte()
	}

	// print when a connection to the client closes along with the error (if any)
	fmt.Printf("Connection close with code of %v and err of: %v\n", code, err)
}
