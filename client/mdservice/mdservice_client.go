package main

import (
	"bufio"
	"fmt"
	"github.com/CPSSD/MDFS/utils"
	"net"
	"os"
	"strconv"
	"strings"
)

func setup(r *bufio.Reader, w *bufio.Writer, thisUser *utils.User) (err error) {
	_, exists := os.Stat(utils.GetUserHome() + "/.client/.user_data")
	if exists != nil { // not exist

		fmt.Println("Make sure the local user dir exist")

		// Make sure the local user dir exists
		err := os.MkdirAll(utils.GetUserHome()+"/.client/", 0777)
		if err != nil {
			return err
		}

		fmt.Println("Notify mdserv of new user")

		// notify mdservice that this is a new user (SENDCODE 10)
		err = w.WriteByte(10) //
		if err != nil {
			return err
		}
		w.Flush()

		fmt.Println("local user setup")

		// local user setup
		utils.GenUserKeys(utils.GetUserHome() + "/.client/.private_key")

		fmt.Println("keys set up")

		err = utils.FileToStruct(utils.GetUserHome()+"/.client/.private_key", &thisUser.Privkey)
		if err != nil {
			return err
		}
		thisUser.Pubkey = &thisUser.Privkey.PublicKey

		fmt.Println("ready to send public key, sending...")

		w.Write([]byte(thisUser.Pubkey.N.String() + "\n"))
		w.Write([]byte(strconv.Itoa(thisUser.Pubkey.E) + "\n"))
		w.Flush()

		fmt.Println("reading uuid")

		uuid, _ := r.ReadString('\n')
		thisUser.Uuid, err = strconv.ParseUint(strings.TrimSpace(uuid), 10, 64)
		if err != nil {
			return err
		}

		fmt.Println("read uuid, store to file")

		err = utils.StructToFile(*thisUser, utils.GetUserHome()+"/.client/.user_data")
		if err != nil {
			return err
		}

		fmt.Println("stored")

		//NOTE: NOT COMPLETE

	} else {

	}

	return err

	// if none exist, will send a code to mdserv to notify as new user,
	// and get a uuid, create userkeys, send pubkey to mdserv

}

func main() {

	// config will be read locally later
	protocol := "tcp"
	socket := "localhost:1994"

	user := "jim"

	// will be filled out in setup, contents of User struct
	// may change slightly to include extra data
	var thisUser utils.User

	conn, _ := net.Dial(protocol, socket)
	defer conn.Close()

	// read and write buffer to the mdserv
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	var sendcode uint8

	// run setup of user.
	// will register with mdservice to get new uuid if a local
	// user file does not exist, and send the user's public key
	// to the mdservice as part of it's registration. The public
	// and private key of this user will be locally accessible when
	// needed by the user client in a location defined in the
	// setup method.
	err := setup(r, w, &thisUser)
	if err != nil {
		panic(err)
	}

	// assume we will always start in the root directory (for safety)
	currentDir := "/"
	reader := bufio.NewReader(os.Stdin)

	// NOTE: after commenting, I have noticed definite code duplication with regards
	// to sending the args for a command and sending the sendcode as well. On the server
	// side this is not as bad as it is here.
	// NOTE: once methods on server side to do with send/request are implemented, comms
	// with the storage node can be brought in from the stnode_client.
	for {

		// print the user's command prompt
		fmt.Print(user + ":" + strconv.FormatUint(thisUser.Uuid, 10) + ":" + currentDir + " >> ")

		// read the next command
		cmd, _ := reader.ReadString('\n')

		// remove trailing newline character before splitting <- didn't notice this
		//                                                       comment before Jacob

		args := strings.Split(strings.TrimSpace(cmd), " ")

		// NOTE: WILL BE PUSHING CASES TO METHODS

		switch args[0] {
		case "":
			continue

		case "ls":
			err := ls(r, w, currentDir, args)
			if err != nil {
				panic(err)
			}

		case "mkdir":
			err := mkdir(w, currentDir, args)
			if err != nil {
				panic(err)
			}

		case "rmdir":
			err := rmdir(r, w, currentDir, args)
			if err != nil {
				panic(err)
			}

		case "cd":
			err := cd(r, w, &currentDir, args)
			if err != nil {
				panic(err)
			}

		case "pwd":
			// no calls to server, just print what we have stored here
			fmt.Print(currentDir + "\n")

		case "exit":
			// leave the program. The server will notice that the client has
			// disconnected and will close the TCP connection on its side
			// without error.
			os.Exit(1)

		case "request":
			// START SENDCODE BLOCK
			sendcode = 5

			err := w.WriteByte(sendcode)
			w.Flush()
			if err != nil {
				panic(err)
			}
			// END SENDCODE BLOCK

			fmt.Printf("Request the file\n")

		case "send":
			// START SENDCODE BLOCK
			sendcode = 6

			err := w.WriteByte(sendcode)
			w.Flush()
			if err != nil {
				panic(err)
			}
			// END SENDCODE BLOCK

			// Send filename to mdserv

			// Get fail if file exists already

			// Get the unid of a storage node if file not exists

			// attempt to send the file (as per stnode_client)
			// INSERT CODE HERE

			// Send success/fail to mdserv to log the file send or not

			// on failure to send a file, print err

		default:

			// you clearly cannot type correctly
			fmt.Println("Unrecognised command")
		}
	}
}

func ls(r *bufio.Reader, w *bufio.Writer, currentDir string, args []string) (err error) {

	// START SENDCODE BLOCK
	err = w.WriteByte(1)
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

	return err

}

func mkdir(w *bufio.Writer, currentDir string, args []string) (err error) {

	// START SENDCODE BLOCK
	err = w.WriteByte(2)
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

	return err
}

func rmdir(r *bufio.Reader, w *bufio.Writer, currentDir string, args []string) (err error) {

	// START SENDCODE BLOCK
	err = w.WriteByte(3)
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

	return err
}

func cd(r *bufio.Reader, w *bufio.Writer, currentDir *string, args []string) (err error) {

	// START SENDCODE BLOCK

	// if the cmd is just "cd", no point telling the server as
	// the result will be no change made to currentDir
	// NOTE: here we do not send more than the first arg as any
	// more than one arg has no effect on results of a cd.
	if len(args) < 2 {
		return err
	}

	err = w.WriteByte(4)
	w.Flush()
	if err != nil {
		panic(err)
	}
	// END SENDCODE BLOCK

	// Send current dir
	w.WriteString(*currentDir + "\n")
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
		*currentDir, _ = r.ReadString('\n')
		*currentDir = strings.TrimSuffix(*currentDir, "\n")
	}

	return err
}
