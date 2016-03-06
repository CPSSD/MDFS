package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	// config will be read locally later
	protocol := "tcp"
	socket := "localhost:1994"
	user := "jim"

	conn, _ := net.Dial(protocol, socket)
	defer conn.Close()

	// read and write buffer to the mdserv
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	var sendcode uint8

	// assume we will always start in the root directory (for safety)
	currentDir := "/"

	// NOTE: after commenting, I have noticed definite code duplication with regards
	// to sending the args for a command and sending the sendcode as well. On the server
	// side this is not as bad as it is here.
	// NOTE: once methods on server side to do with send/request are implemented, comms
	// with the storage node can be brought in from the stnode_client.
	for {

		// get a new reader, perhaps shouldn't do this everytime, although maybe
		// this will stop user commands from being entered while waiting for results
		// of commands from the mdservice which could be a good thing.
		reader := bufio.NewReader(os.Stdin)

		// print the user's command prompt
		fmt.Print(user + ":" + currentDir + " >> ")

		// read the next command
		cmd, _ := reader.ReadString('\n')

		// remove trailing newline character before splitting <- didn't notice this
		//														 comment before Jacob

		args := strings.Split(strings.TrimSpace(cmd), " ")

		switch args[0] {
		case "":
			continue

		case "ls":
			// START SENDCODE BLOCK
			sendcode = 1

			err := w.WriteByte(sendcode)
			w.Flush()
			if err != nil {
				panic(err)
			}
			// END SENDCODE BLOCK

			// Send current dir
			w.WriteString(currentDir + "\n")
			w.Flush()

			// send the length of args
			err = w.WriteByte(uint8(len(args)))
			w.Flush()
			if err != nil {
				panic(err)
			}

			// write each arg (if there are any) seperately so that the server
			// can deal with them as per it's loop
			for i := 1; i < len(args); i++ {

				// simple sending of args
				w.WriteString(args[i] + "\n")
				w.Flush()
			}

			// read to whitespace? Check corresponding write on the server side
			msg, _ := r.ReadString(' ')

			// split results by commas
			files := strings.Split(msg, ",")

			// remove the last newline
			msg = strings.TrimSuffix(msg, "\n")

			// iterate over each result of the comma separated msg
			for n, file := range files {

				// don't print the last one? this is hacky code, needs cleaning on
				// both client and server side, make sure to test changes so as not
				// to break pieces
				if n != len(files)-1 {

					// newline for each piece of the result
					fmt.Println(file)
				}
			}

		case "mkdir":
			// START SENDCODE BLOCK
			sendcode = 2

			err := w.WriteByte(sendcode)
			w.Flush()
			if err != nil {
				panic(err)
			}
			// END SENDCODE BLOCK

			// Send current dir
			w.WriteString(currentDir + "\n")
			w.Flush()

			// send len args for mkdir
			err = w.WriteByte(uint8(len(args)))
			w.Flush()
			if err != nil {
				panic(err)
			}

			// send each arg (if it exists)
			for i := 1; i < len(args); i++ {

				w.WriteString(args[i] + "\n")
				w.Flush()
			}

		case "rmdir":
			// START SENDCODE BLOCK
			sendcode = 3

			err := w.WriteByte(sendcode)
			w.Flush()
			if err != nil {
				panic(err)
			}
			// END SENDCODE BLOCK

			// Send current dir
			w.WriteString(currentDir + "\n")
			w.Flush()

			// send len args
			err = w.WriteByte(uint8(len(args)))
			w.Flush()
			if err != nil {
				panic(err)
			}

			// send each arg (if it exists)
			for i := 1; i < len(args); i++ {

				w.WriteString(args[i] + "\n")
				w.Flush()
			}

		case "cd":
			// START SENDCODE BLOCK
			sendcode = 4

			// if the cmd is just "cd", no point telling the server as
			// the result will be no change made to currentDir
			// NOTE: here we do not send more than the first arg as any
			// more than one arg has no effect on results of a cd.
			if len(args) < 2 {
				continue
			}

			err := w.WriteByte(sendcode)
			w.Flush()
			if err != nil {
				panic(err)
			}
			// END SENDCODE BLOCK

			// Send current dir
			w.WriteString(currentDir + "\n")
			w.Flush()

			// Send target dir
			w.WriteString(args[1] + "\n")
			w.Flush()

			// get response from server if is a dir
			isDir, _ := r.ReadByte()
			if isDir == 1 {

				fmt.Println("Not a directory")

			} else { // success!

				// get the new currentDir and clean up the end
				currentDir, _ = r.ReadString('\n')
				currentDir = strings.TrimSuffix(currentDir, "\n")
			}

		case "pwd":
			// no calls to server, just print what we have stored here
			fmt.Print(currentDir + "\n")

		case "exit":
			// leave the program. The server will notice that the client has
			// disconnected and will close the TCP connection on its side
			// without error.
			os.Exit(1)

		case "send":
			// START SENDCODE BLOCK
			sendcode = 5

			err := w.WriteByte(sendcode)
			w.Flush()
			if err != nil {
				panic(err)
			}
			// END SENDCODE BLOCK

			fmt.Printf("Send the file\n")

		case "request":
			// START SENDCODE BLOCK
			sendcode = 6

			err := w.WriteByte(sendcode)
			w.Flush()
			if err != nil {
				panic(err)
			}
			// END SENDCODE BLOCK

			fmt.Printf("Request the file\n")

		default:

			// you clearly cannot type correctly
			fmt.Println("Unrecognised command")
		}
	}
}
