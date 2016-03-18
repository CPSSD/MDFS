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

	fmt.Println("Please enter your username:\n")
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	uname, _ := reader.ReadString('\n')
	uname = strings.TrimSpace(uname)

	thisUser.Uname = uname
	_, exists := os.Stat(utils.GetUserHome() + "/.mdfs/client/" + uname + "/.user_data")
	if exists != nil { // not exist

		// Make sure the local user dir exists
		err := os.MkdirAll(utils.GetUserHome()+"/.mdfs/client/"+uname+"/files", 0777)
		if err != nil {
			return err
		}

		// notify mdservice that this is a new user (SENDCODE 10)
		err = w.WriteByte(10) //
		if err != nil {
			return err
		}
		w.Flush()

		// local user setup
		utils.GenUserKeys(utils.GetUserHome() + "/.mdfs/client/" + uname + "/.private_key")

		err = utils.FileToStruct(utils.GetUserHome()+"/.mdfs/client/"+uname+"/.private_key", &thisUser.Privkey)
		if err != nil {
			return err
		}
		thisUser.Pubkey = &thisUser.Privkey.PublicKey

		// send username and keys
		w.WriteString(uname + "\n")
		w.Write([]byte(thisUser.Pubkey.N.String() + "\n"))
		w.Write([]byte(strconv.Itoa(thisUser.Pubkey.E) + "\n"))
		w.Flush()

		uuid, _ := r.ReadString('\n')
		thisUser.Uuid, err = strconv.ParseUint(strings.TrimSpace(uuid), 10, 64)
		if err != nil {
			return err
		}

		err = utils.StructToFile(*thisUser, utils.GetUserHome()+"/.mdfs/client/"+uname+"/.user_data")
		if err != nil {
			return err
		}

		//NOTE: NOT COMPLETE

	} else {

		err = utils.FileToStruct(utils.GetUserHome()+"/.mdfs/client/"+uname+"/.user_data", &thisUser)
		w.WriteByte(9)

	}

	return err

	// if none exist, will send a code to mdserv to notify as new user,
	// and get a uuid, create userkeys, send pubkey to mdserv

}

func main() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("----------------------------------------------------------")
	fmt.Println("Please enter the location of the metadata service you wish to connect to (eg: 192.168.0.10)\n")
	fmt.Print("Address: ")
	host, _ := reader.ReadString('\n')
	fmt.Println("----------------------------------------------------------")
	host = strings.TrimSpace(host)

	fmt.Println("Please enter the port of the metadata service to connect to.\n")
	fmt.Print("Port: ")
	port, _ := reader.ReadString('\n')
	fmt.Println("----------------------------------------------------------")
	port = strings.TrimSpace(port)

	protocol := "tcp"
	addr := host + ":" + port

	// will be filled out in setup, contents of User struct
	// may change slightly to include extra data
	var thisUser utils.User

	fmt.Println("Connecting to metadata service...")
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Println("No mdserv available through " + protocol + " connection at " + addr)
		os.Exit(0)
	}
	fmt.Println("Connection established.")

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

	fmt.Println("----------------------------------------------------------")
	err = setup(r, w, &thisUser)
	if err != nil {
		panic(err)
	}
	fmt.Println("----------------------------------------------------------")

	idStr := strconv.FormatUint(thisUser.Uuid, 10)
	w.WriteString(idStr + "\n")
	w.Flush()

	// assume we will always start in the root directory (for safety)
	currentDir := "/"

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

		case "rm":
			err := rm(r, w, currentDir, args)
			if err != nil {
				panic(err)
			}

		case "pwd":
			// no calls to server, just print what we have stored here
			fmt.Print(currentDir + "\n")

		case "create-group":
			err := createGroup(r, w, args, &thisUser)
			if err != nil {
				panic(err)
			}

		case "group-add":
			err := groupAdd(r, w, args, &thisUser)
			if err != nil {
				panic(err)
			}

		case "group-remove":
			err := groupRemove(r, w, args, &thisUser)
			if err != nil {
				panic(err)
			}

		case "group-ls":
			err := groupLs(r, w, args, &thisUser)
			if err != nil {
				panic(err)
			}

		case "delete-group":
			err := deleteGroup(r, w, args, &thisUser)
			if err != nil {
				panic(err)
			}

		case "list-groups":
			err := listGroups(r, w, args, &thisUser)
			if err != nil {
				panic(err)
			}

		case "permit":
			err := permit(currentDir, r, w, args, &thisUser)
			if err != nil {
				panic(err)
			}

		case "deny":
			err := deny(currentDir, r, w, args, &thisUser)
			if err != nil {
				panic(err)
			}

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

	verbose := 0
	verboseMod := 0
	// END SENDCODE BLOCK
	if len(args) > 1 && args[1] == "-V" {
		verbose = 1
		verboseMod = 1
	}
	// Send current dir
	w.WriteString(currentDir + "\n")
	w.Flush()

	// send the length of args
	err = w.WriteByte(uint8(len(args)))
	w.Flush()
	if err != nil {
		panic(err)
	}

	if len(args) == 2 && verbose == 1 {

		w.WriteByte(1) // verbose
		w.Flush()

		inFiles, _ := r.ReadByte()
		numFiles := int(inFiles)

		for i := 0; i < numFiles; i++ {
			file, _ := r.ReadString('\n')
			fmt.Print(file)
		}
		return nil

	} else if verbose == 1 {

		w.WriteByte(1) // verbose
		w.Flush()

	} else if len(args) == 1 {

		w.WriteByte(2)
		w.Flush()
		inFiles, _ := r.ReadByte()
		numFiles := int(inFiles)

		for i := 0; i < numFiles; i++ {
			file, _ := r.ReadString('\n')
			fmt.Print(file)
		}
		return nil

	} else {

		w.WriteByte(2)
		w.Flush()

	}

	// write each arg (if there are any) seperately so that the server
	// can deal with them as per it's loop
	for i := 1 + verboseMod; i < len(args); i++ {

		// simple sending of args
		w.WriteString(args[i] + "\n")
		w.Flush()

		inFiles, _ := r.ReadByte()
		numFiles := int(inFiles)

		for i := 0; i < numFiles; i++ {
			file, _ := r.ReadString('\n')
			fmt.Print(file)

		}

		if i != len(args)-1 && numFiles != 0 {
			fmt.Println()
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
func rm(r *bufio.Reader, w *bufio.Writer, currentDir string, args []string) (err error) {

	// START SENDCODE BLOCK
	err = w.WriteByte(7)
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

	} else if isDir == 2 {

		fmt.Println("Permission denied")

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

		fmt.Println("Sending: \"" + filepath + "\"")

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
	// Send filename to mdserv
	w.WriteString(path.Base(filepath) + "\n")

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
			fmt.Println("Invalid argument option: \"" + args[2] + "\"")
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
		fmt.Println("File successfully protected")

	} else {

		w.WriteByte(3) // normal continue
		w.Flush()
	}

	// get hash of the file to send to the stnode and mdserv
	hash, err := utils.ComputeMd5(filepath)
	if err != nil {
		panic(err)
	}

	//checksum := hex.EncodeToString(hash)

	// Send hash to mdserv
	err = utils.WriteHash(w, hash)
	if err != nil {
		return err
	}

	// See are there stnodes available
	avail, _ := r.ReadByte()

	var conns net.Conn

	for avail != 2 {

		// Get details of a storage node if file not exists
		protocol, _ := r.ReadString('\n')
		nAddress, _ := r.ReadString('\n')

		protocol = strings.TrimSpace(protocol)
		nAddress = strings.TrimSpace(nAddress)

		// connect to stnode
		conns, err = net.Dial(protocol, nAddress)
		if err != nil {

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

	defer conns.Close()

	// create a read and write buffer for the connection
	ws := bufio.NewWriter(conns)

	// tell the stnode we are sending a file
	err = ws.WriteByte(2)
	if err != nil {
		return err
	}

	// send hash to stnode
	err = utils.WriteHash(ws, hash)
	if err != nil {
		return err
	}

	// send file to stnode
	utils.SendFile(conns, ws, filepath)

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

	fmt.Println("Requesting: " + path.Join(currentDir, args[1]))

	success, _ := r.ReadByte()
	if success == 3 {

		fmt.Println("You do not have permission to request this file")
		return nil

	} else if success != 4 {

		fmt.Println("Invalid file request")
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

	output := utils.GetUserHome() + "/.mdfs/client/" + thisUser.Uname + "/files/" + path.Base(args[1])

	if protected {

		encrypFile := os.TempDir() + "/" + path.Base(args[1])
		utils.ReceiveFile(conn, rs, encrypFile)

		err = utils.DecryptFile(encrypFile, output, *thisUser)
		if err != nil {

			fmt.Println("Your key does not match the lock for this file")
			return nil
		}
		fmt.Println("Protected file successfully unlocked")

	} else {

		utils.ReceiveFile(conn, rs, output)

	}

	fmt.Println("Successfully received file")
	fmt.Println("File stored at: " + output)

	return
}

func createGroup(r *bufio.Reader, w *bufio.Writer, args []string, thisUser *utils.User) (err error) {

	if len(args) == 1 {
		fmt.Println("No arguments for call to create-group.")
		return nil
	}

	err = w.WriteByte(20)
	w.Flush()
	if err != nil {
		return err
	}

	// send the length of args
	err = w.WriteByte(uint8(len(args)))
	w.Flush()
	if err != nil {
		return err
	}

	for i := 1; i < len(args); i++ {

		// send group to create
		w.WriteString(args[i] + "\n")
		w.Flush()

		// get group id to display to user
		gid, _ := r.ReadString('\n')
		fmt.Println("Created group \"" + args[i] + "\" with id: " + strings.TrimSpace(gid))
	}
	// send the group to create

	return err
}

func groupAdd(r *bufio.Reader, w *bufio.Writer, args []string, thisUser *utils.User) (err error) {

	if len(args) < 3 {
		fmt.Println("Not enough arguments for call to groupadd:")
		fmt.Println("Format should be: group-add GID UUID1 UUID2 ... UUIDN")
		return nil
	}

	err = w.WriteByte(21)
	w.Flush()
	if err != nil {
		return err
	}

	// send the length of args
	err = w.WriteByte(uint8(len(args)))
	w.Flush()
	if err != nil {
		return err
	}

	w.WriteString(args[1] + "\n")
	w.Flush()

	// get success (1) or fail (2)
	success, _ := r.ReadByte()
	if success != 1 {
		fmt.Println("You cannot add users to this group. Are you the owner?")
		return err
	}

	for i := 2; i < len(args); i++ {

		// send uuid to add
		w.WriteString(args[i] + "\n")
		w.Flush()
	}

	// get uuids added
	result, _ := r.ReadString('\n')
	result = strings.TrimSuffix(strings.TrimSpace(result), ",")

	fmt.Println("Added users: " + result + " to group " + args[1])

	return err
}

func groupRemove(r *bufio.Reader, w *bufio.Writer, args []string, thisUser *utils.User) (err error) {

	if len(args) < 3 {
		fmt.Println("Not enough arguments for call to group-remove:")
		fmt.Println("Format should be: group-remove GID UUID1 UUID2 ... UUIDN")
		return nil
	}

	err = w.WriteByte(22)
	w.Flush()
	if err != nil {
		return err
	}

	// send the length of args
	err = w.WriteByte(uint8(len(args)))
	w.Flush()
	if err != nil {
		return err
	}

	w.WriteString(args[1] + "\n")
	w.Flush()

	// get success (1) or fail (2)
	success, _ := r.ReadByte()
	if success != 1 {
		fmt.Println("You cannot remove users from this group. Are you the owner?")
		return err
	}

	for i := 2; i < len(args); i++ {

		// send uuid to remove
		w.WriteString(args[i] + "\n")
		w.Flush()
	}

	// get uuids removed
	result, _ := r.ReadString('\n')
	result = strings.TrimSuffix(strings.TrimSpace(result), ",")

	fmt.Println("Removed users: " + result + " from group " + args[1])

	return err
}

func groupLs(r *bufio.Reader, w *bufio.Writer, args []string, thisUser *utils.User) (err error) {

	if len(args) != 2 {
		fmt.Println("Wrong number of arguments for call to group-ls:")
		fmt.Println("Format should be: group-ls GID")
		return nil
	}

	err = w.WriteByte(23)
	w.Flush()
	if err != nil {
		return err
	}

	// send the group id
	w.WriteString(args[1] + "\n")
	w.Flush()

	// get success (1) or fail (2)
	success, _ := r.ReadByte()
	if success != 1 {
		fmt.Println("The group does not exist")
		return err
	}

	// get uuids of the group
	result, _ := r.ReadString('\n')
	result = strings.TrimSuffix(strings.TrimSpace(result), ",")

	fmt.Println("Members of Group " + args[1] + ": " + result)

	return err
}

func deleteGroup(r *bufio.Reader, w *bufio.Writer, args []string, thisUser *utils.User) (err error) {

	if len(args) < 2 {
		fmt.Println("Not enough arguments for call to group-remove:")
		fmt.Println("Format should be: delete-group GID")
		return nil
	}

	err = w.WriteByte(24)
	w.Flush()
	if err != nil {
		return err
	}

	w.WriteString(args[1] + "\n")
	w.Flush()

	// get success (1) or fail (2)
	success, _ := r.ReadByte()
	if success != 1 {
		fmt.Println("You cannot remove this group. Does it exist/are you the owner?")
		return err
	}

	fmt.Println("Removed: " + args[1])

	return err
}

func listGroups(r *bufio.Reader, w *bufio.Writer, args []string, thisUser *utils.User) (err error) {

	w.WriteByte(25)
	w.Flush()

	verbose := false

	if len(args) == 1 { // sendcode 25

		w.WriteByte(3)
		w.Flush()

	} else {

		switch args[1] {
		case "-m": // sendcode 26
			fmt.Println("You are a member of the following groups:")
			w.WriteByte(1)
			w.Flush()

		case "-o": // sendcode 27
			fmt.Println("You are the owner of the following groups:")
			w.WriteByte(2)
			w.Flush()

		case "-mV": // sendcode 26
			fmt.Println("You are a member of the following groups:")
			w.WriteByte(1)
			w.Flush()
			verbose = true

		case "-oV": // sendcode 27
			fmt.Println("You are the owner of the following groups:")
			w.WriteByte(2)
			w.Flush()
			verbose = true

		case "-V": // sendcode 27
			fmt.Println("You are the owner of the following groups:")
			w.WriteByte(3)
			w.Flush()
			verbose = true

		default:
			w.WriteByte(4)
			w.Flush()
			fmt.Println("Not valid arguments to list-groups:")
			fmt.Println("Format should be one of:")
			fmt.Println("list-groups -m")
			fmt.Println("list-groups -mV")
			fmt.Println("list-groups -o")
			fmt.Println("list-groups -oV")
			fmt.Println("list-groups")
			return nil
		}
	}

	for success, _ := r.ReadByte(); success != 2; success, _ = r.ReadByte() {

		group, _ := r.ReadString('\n')
		groupArr := strings.Split(strings.TrimSuffix(strings.TrimSpace(group), ","), ",")

		fmt.Print("Name: " + groupArr[0] + "\tID: " + groupArr[1])

		if verbose {
			fmt.Print("\tMembers: ")
			for i, v := range groupArr {
				if i > 1 {
					fmt.Print(v + ", ")
				}
			}
		}
		fmt.Println()
	}

	return
}

func permit(currentDir string, r *bufio.Reader, w *bufio.Writer, args []string, thisUser *utils.User) (err error) {

	if len(args) < 3 {
		fmt.Println("Not enough arguments for call to permit")
	}

	w.WriteByte(8)
	w.Flush()

	w.WriteString(currentDir + "\n")
	w.Flush()

	switch args[1] {
	case "-g":
		w.WriteString(args[1] + "\n")
	case "-w":
		w.WriteString(args[1] + "\n")
	default:
		w.WriteString("INV" + "\n")
		w.Flush()
		fmt.Println("Invalid switch for call to permit; valid switches are -g or -w")
		return nil
	}
	w.Flush()

	err = w.WriteByte(uint8(len(args)))
	w.Flush()
	if err != nil {
		panic(err)
	}

	w.WriteString(args[2] + "\n")
	w.Flush()

	for i := 3; i < len(args); i++ {
		w.WriteString(args[i] + "\n")
		w.Flush()
	}

	// 8
	return nil
}

func deny(currentDir string, r *bufio.Reader, w *bufio.Writer, args []string, thisUser *utils.User) (err error) {

	if len(args) < 3 {
		fmt.Println("Not enough arguments for call to permit")
	}

	w.WriteByte(9)
	w.Flush()

	w.WriteString(currentDir + "\n")
	w.Flush()

	switch args[1] {
	case "-g":
		w.WriteString(args[1] + "\n")
	case "-w":
		w.WriteString(args[1] + "\n")
	default:
		w.WriteString("INV" + "\n")
		w.Flush()
		fmt.Println("Invalid switch for call to deny; valid switches are -g or -w")
		return nil
	}
	w.Flush()

	err = w.WriteByte(uint8(len(args)))
	w.Flush()
	if err != nil {
		panic(err)
	}

	w.WriteString(args[2] + "\n")
	w.Flush()

	for i := 3; i < len(args); i++ {
		w.WriteString(args[i] + "\n")
		w.Flush()
	}

	// 9
	return
}
