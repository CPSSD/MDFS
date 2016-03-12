package main

import (
	"bufio"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"github.com/CPSSD/MDFS/utils"
	"math/big"
	"net"
	"os"
	"path"
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

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Please enter your username and hit enter: ")
		uname, _ := reader.ReadString('\n')
		thisUser.Uname = strings.TrimSpace(uname)

		// local user setup
		utils.GenUserKeys(utils.GetUserHome() + "/.client/.private_key")

		fmt.Println("keys set up")

		err = utils.FileToStruct(utils.GetUserHome()+"/.client/.private_key", &thisUser.Privkey)
		if err != nil {
			return err
		}
		thisUser.Pubkey = &thisUser.Privkey.PublicKey

		fmt.Println("ready to send public key, sending...")

		// send username and keys
		w.WriteString(uname)
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

		err = utils.FileToStruct(utils.GetUserHome()+"/.client/.user_data", &thisUser)
	}

	return err

	// if none exist, will send a code to mdserv to notify as new user,
	// and get a uuid, create userkeys, send pubkey to mdserv

}

func main() {

	// config will be read locally later
	protocol := "tcp"
	socket := "localhost:1994"

	// will be filled out in setup, contents of User struct
	// may change slightly to include extra data
	var thisUser utils.User

	conn, err := net.Dial(protocol, socket)
	if err != nil {
		fmt.Println("No mdserv available through " + protocol + " connection at " + socket)
		os.Exit(0)
	}
	defer conn.Close()

	// read and write buffer to the mdserv
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	//var sendcode uint8

	// run setup of user.
	// will register with mdservice to get new uuid if a local
	// user file does not exist, and send the user's public key
	// to the mdservice as part of it's registration. The public
	// and private key of this user will be locally accessible when
	// needed by the user client in a location defined in the
	// setup method.
	err = setup(r, w, &thisUser)
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
		fmt.Print(thisUser.Uname + ":" + strconv.FormatUint(thisUser.Uuid, 10) + ":" + currentDir + " >> ")

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

		case "request":
			err := request(r, w, currentDir, args, &thisUser)
			if err != nil {
				panic(err)
			}

		case "send":
			err := send(r, w, currentDir, args, &thisUser)
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
			os.Exit(0)

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

func send(r *bufio.Reader, w *bufio.Writer, currentDir string, args []string, thisUser *utils.User) (err error) {

	if len(args) < 2 {
		return err
	}

	fmt.Println("Base of file :" + path.Base(args[1]))

	// Format the file to send (absolute or relative)
	filepath := ""
	if path.IsAbs(args[1]) { // if we are trying to send an absolute filepath

		filepath = args[1]

	} else { // we must be sending a relative filepath, so calculate the path

		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		filepath = path.Join(wd, args[1])
	}

	// ensure the file is existant locally
	src, err := os.Stat(filepath)
	if err != nil {

		fmt.Println("\"" + filepath + "\" does not exist")
		return nil

	} else if src.IsDir() {

		fmt.Println("\"" + filepath + "\" is a directory, not a file")
		return nil

	} else {

		fmt.Println("Filepath to send is: \"" + filepath + "\"")

	}

	// START SENDCODE BLOCK - tell mdserv we are sending a file
	err = w.WriteByte(6)
	w.Flush()
	if err != nil {
		panic(err)
	}
	// END SENDCODE BLOCK

	// Send current dir
	w.WriteString(currentDir + "\n")
	fmt.Println("Send: current dir: " + currentDir)
	// Send filename to mdserv
	w.WriteString(path.Base(filepath) + "\n")
	fmt.Println("Send: filename: " + filepath)

	w.Flush()

	// Get fail if file exists already
	exists, _ := r.ReadByte()
	if exists == 1 {

		fmt.Println("Bad send request, check file name again")
		return nil
	} else if exists != 2 {

		fmt.Println("Bad response from mdserv")
		return nil
	}

	// Encryption or not?
	if len(args) >= 3 {
		if args[2] != "--protect" && args[2] != "-p" {
			fmt.Println("Invalid command option")
			w.WriteByte(2)
			w.Flush()
			return nil
		}

		w.WriteByte(1) // tell mdserv we are encrypting
		w.Flush()

		users := []utils.User{*thisUser}

		// send len of args
		w.WriteByte(uint8(len(args) - 3))
		w.Flush()

		for i := 3; i < len(args); i++ {

			w.WriteString(args[i] + "\n")
			w.Flush()

			fmt.Println("Sent: " + args[i])

			uuid, _ := r.ReadString('\n')
			uuid = strings.TrimSpace(uuid)

			if uuid != "INV" {

				var newUser utils.User

				newUser.Uuid, _ = strconv.ParseUint(uuid, 10, 64)

				// receive the public key for the new user

				pubKN, _ := r.ReadString('\n')
				pubKE, _ := r.ReadString('\n')

				newUser.Pubkey = &rsa.PublicKey{N: big.NewInt(0)}
				newUser.Pubkey.N.SetString(strings.TrimSpace(pubKN), 10)
				newUser.Pubkey.E, err = strconv.Atoi(strings.TrimSpace(pubKE))

				users = append(users, newUser)
			}
		}

		encrypFile := os.TempDir() + "/" + path.Base(filepath)

		err = utils.EncryptFile(filepath, encrypFile, users...)
		if err != nil {
			return err
		}

		filepath = encrypFile
		fmt.Println("Encrypted file to: " + filepath + " from " + encrypFile)

	} else {

		w.WriteByte(3) // normal continue
		w.Flush()
	}

	fmt.Println("Computing hash")
	// get hash of the file to send to the stnode and mdserv
	hash, err := utils.ComputeMd5(filepath)
	if err != nil {
		panic(err)
	}

	checksum := hex.EncodeToString(hash)
	fmt.Println("Computed hash: " + checksum)

	fmt.Println("File does not exist, sending hash")

	// Send hash to mdserv
	fmt.Println("Sending hash to mdserv: " + checksum)
	err = utils.WriteHash(w, hash)
	if err != nil {
		return err
	}

	fmt.Println("Hash sent, see are there stnodes available")

	// See are there stnodes available
	avail, _ := r.ReadByte()

	var conn net.Conn

	for avail != 2 {

		// Get details of a storage node if file not exists
		protocol, _ := r.ReadString('\n')
		nAddress, _ := r.ReadString('\n')

		protocol = strings.TrimSpace(protocol)
		nAddress = strings.TrimSpace(nAddress)

		fmt.Println("protocol: " + protocol + ", address: " + nAddress)

		// connect to stnode
		conn, err = net.Dial(protocol, nAddress)
		if err != nil {

			fmt.Println("Error connecting to stnode")
			w.WriteByte(1)
			w.Flush()

		} else { // successful connection

			avail = 3
			w.WriteByte(2)
			w.Flush()
			break
		}

		avail, _ = r.ReadByte()
	}

	if avail != 3 {
		fmt.Println("There were no stnodes available")
		return nil
	}

	defer conn.Close()

	// create a read and write buffer for the connection
	ws := bufio.NewWriter(conn)

	// tell the stnode we are sending a file
	err = ws.WriteByte(2)
	if err != nil {
		return err
	}

	// send hash to stnode
	fmt.Println("Sending hash to stnode: " + checksum)
	err = utils.WriteHash(ws, hash)
	if err != nil {
		return err
	}

	// send file to stnode
	fmt.Println("Sending: " + filepath)
	utils.SendFile(conn, ws, filepath)

	// Send success/fail to mdserv to log the file send or not
	w.WriteByte(1)
	w.Flush()

	fmt.Println("Successfully sent file")
	// on failure to send a file, print err
	return err
}

func request(r *bufio.Reader, w *bufio.Writer, currentDir string, args []string, thisUser *utils.User) (err error) {

	// should have args format:
	// request [remote filename] [local filename]
	// currently:
	// request [remote filename]
	// Local filename can be a relative or absolute path

	if len(args) < 2 {
		return err
	}

	// START SENDCODE BLOCK
	err = w.WriteByte(5)
	w.Flush()
	if err != nil {
		panic(err)
	}
	// END SENDCODE BLOCK

	// Send current dir
	w.WriteString(currentDir + "\n")
	w.Flush()

	// Format the file to send
	w.WriteString(args[1] + "\n")
	w.Flush()

	success, _ := r.ReadByte()
	if success != 3 {

		fmt.Printf("Invalid file request with response: %v\n", success)
		return nil
	}

	protected := false
	enc, _ := r.ReadByte()
	if enc == 1 {
		protected = true
	}

	success, _ = r.ReadByte()
	if success != 2 {

		fmt.Println("No stnodes for your file")
		return nil
	}

	hash, _ := r.ReadString('\n')
	protocol, _ := r.ReadString('\n')
	nAddress, _ := r.ReadString('\n')

	hash = strings.TrimSpace(hash)
	fmt.Println("Received hash: " + hash)
	protocol = strings.TrimSpace(protocol)
	nAddress = strings.TrimSpace(nAddress)

	conn, err := net.Dial(protocol, nAddress)
	if err != nil {
		fmt.Println("Error connecting to stnode")
	}

	ws := bufio.NewWriter(conn)
	rs := bufio.NewReader(conn)

	ws.WriteByte(1)

	bytehash, err := hex.DecodeString(hash)
	if err != nil {
		return err
	}

	err = utils.WriteHash(ws, bytehash)
	if err != nil {
		return err
	}

	success, _ = rs.ReadByte()
	if success != 3 {

		fmt.Println("File cannot be found on stnode")
		return err
	}

	output := utils.GetUserHome() + "/.client/" + path.Base(args[1])

	if protected {

		encrypFile := os.TempDir() + "/" + path.Base(args[1])
		utils.ReceiveFile(conn, rs, encrypFile)

		err = utils.DecryptFile(encrypFile, output, *thisUser)
		if err != nil {
			return err
		}

	} else {

		utils.ReceiveFile(conn, rs, output)

	}

	fmt.Println("File exists")

	return
}
