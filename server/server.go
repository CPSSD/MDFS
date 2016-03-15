package server

import (
	"bufio"
	"crypto/rsa"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
)

type Server struct {
	conf config.Configuration
}

type StorageNode struct {
	Server // anonymous field of type Server
}

type MDService struct {
	userDB   *bolt.DB
	stnodeDB *bolt.DB
	groupDB  *bolt.DB
	Server   // anonymous field of type Server
}

// the Server interface
type TCPServer interface {
	parseConfig()
	setup() error
	finish()
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

func (st StorageNode) getUnid() string {
	return st.conf.Unid
}

func (st *StorageNode) setUnid(unid string) (err error) {

	fmt.Println(st.getUnid())
	st.conf.Unid = unid
	fmt.Println(st.getUnid())
	err = config.SetConfiguration(st.conf, utils.GetUserHome()+"/.mdfs/stnode/.stnode_conf.json")
	if err != nil {
		return err
	}
	return err
}

// StorageNode methods
func (st *StorageNode) parseConfig() {
	st.conf = config.ParseConfiguration(utils.GetUserHome() + "/.mdfs/stnode/.stnode_conf.json")
}

func (st *StorageNode) setup() (err error) {

	fmt.Println(st.getUnid())

	if st.getUnid() == "-1" {
		// stnode will register with the mdserv here
		protocol := "tcp"
		socket := "localhost:1994"

		fmt.Println("Connecting to mdserv")
		conn, _ := net.Dial(protocol, socket)
		defer conn.Close()

		// read and write buffer to the mdserv
		r := bufio.NewReader(conn)
		w := bufio.NewWriter(conn)

		var sendcode uint8
		sendcode = 11

		fmt.Println("Registering with mdserv")
		// tell the mdserv that we are connecting to register this stnode
		w.WriteByte(sendcode)
		w.Flush()

		fmt.Println("Sending connection details to mdserv")
		// tell the mdserv the connection details for this stnode
		w.WriteString(st.getProtocol() + "\n")
		w.WriteString(st.getHost() + ":" + st.getPort() + "\n")
		w.Flush()

		fmt.Println("Waiting to receive unid")
		// get the unid for this stnode
		unid, _ := r.ReadString('\n')
		unid = strings.TrimSpace(unid)

		st.setUnid(unid)

		fmt.Println("Received unid: " + unid)
	} else {

		fmt.Println("Stnode has UNID: " + st.getUnid())
	}

	return err
}

func (st *StorageNode) finish() {

}

// MDService methods
// initialise its memeber variable with values from config file
func (md *MDService) parseConfig() {
	md.conf = config.ParseConfiguration(utils.GetUserHome() + "/.mdfs/mdservice/.mdservice_conf.json")
}

// open user and stnode db
func (md *MDService) setup() (err error) {

	// init the boltdb if it is not existant already
	// one for users, one for stnodes
	fmt.Println("This is a metadata service, opening DB's")
	md.userDB, err = bolt.Open(md.getPath()+".userDB.db", 0700, nil)
	if err != nil {
		return err
	}

	err = md.userDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	err = md.userDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("groups"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	fmt.Println("Set up user db, containing user bucket and group bucket")

	md.stnodeDB, err = bolt.Open(md.getPath()+".stnodeDB.db", 0700, nil)
	if err != nil {
		return err
	}

	err = md.stnodeDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("stnodes"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	fmt.Println("Set up stnode db, containing stnode bucket")

	return err
}

// close user and stnode db
func (md *MDService) finish() {

	fmt.Println("Ready to close dbs")
	md.userDB.Close()
	md.stnodeDB.Close()
	fmt.Println("Closed dbs")
}

// checks request code and calls corresponding function
func (st StorageNode) handleCode(code uint8, conn net.Conn, r *bufio.Reader, w *bufio.Writer) {

	switch code {
	case 1: // client is requesting a file
		hash := utils.ReadHashAsString(r)
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
			w.Flush()

			// send the file
			utils.SendFile(conn, w, fp)
		} else {
			sendcode = 4
			err := w.WriteByte(sendcode) // let client know
			if err != nil {
				panic(err)
			}
			w.Flush()
		}

	case 2: // receive file from client
		hash := utils.ReadHashAsString(r)
		output := st.getPath() + hash
		utils.ReceiveFile(conn, r, output)

		fmt.Println("md5 checksum of file is: " + hash)
	}
	conn.Close()
}

func (md MDService) handleCode(code uint8, conn net.Conn, r *bufio.Reader, w *bufio.Writer) {

	// switch statement for commands
	// commands work but are not yet optimized, possible code duplication occurs,
	// although moving some code to a function to reduce duping may not be worth it
	// if it is only one or two lines.
	switch code {
	case 1:
		fmt.Println("In ls")
		err := ls(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Exited ls")

	case 2:
		fmt.Println("In mkdir")
		err := mkdir(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin mkdir")

	case 3:
		fmt.Println("In rmdir")
		err := rmdir(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin rmdir")

	case 4: // cd
		fmt.Println("In cd")
		err := cd(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin cd")

	case 5: // request
		fmt.Println("In request")
		err := request(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin request")

	case 6: // send
		fmt.Println("In send")
		err := send(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin send")

	case 7: // rm
		fmt.Println("In rm")
		err := rm(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin rm")

	case 10: // setup new user
		fmt.Println("In user setup")
		err := newUser(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin user setup")

	case 11: // setup new storage node
		fmt.Println("In stnode setup")
		err := newStnode(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin stnode setup")

	case 20:
		fmt.Println("In createGroup")
		err := createGroup(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin createGroup")

	case 21:
		fmt.Println("In groupAdd")
		err := groupAdd(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin groupAdd")

	case 22:
		fmt.Println("In groupRemove")
		err := groupRemove(conn, r, w, &md)
		if err != nil {
			panic(err)
		}
		fmt.Println("Fin groupRemove")
	}
}

func itob(v uint64) []byte { // convert uint64 to byte array
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	return b
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
	// stnode will register with mdservice here
	err := in.setup()
	if err != nil {
		panic(err)
	}
	defer in.finish()

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

	// mdservice closes db here etc
	in.finish()
}

func createFile(fileout, hash, unid string, protected bool) error {

	var tmpFileDesc utils.FileDesc

	tmpFileDesc.Hash = hash
	tmpFileDesc.Stnode = unid
	tmpFileDesc.Protected = protected

	return utils.StructToFile(tmpFileDesc, fileout)
}

func getFile(fileout string) (hash, unid string, protected bool, err error) {

	var tmpFileDesc utils.FileDesc

	err = utils.FileToStruct(fileout, &tmpFileDesc)
	hash = tmpFileDesc.Hash
	unid = tmpFileDesc.Stnode
	protected = tmpFileDesc.Protected
	return
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
		r = bufio.NewReader(conn)
		w = bufio.NewWriter(conn)

		// wait to read the next code from the client
		code, err = r.ReadByte()
	}

	// print when a connection to the client closes along with the error (if any)
	fmt.Printf("Connection close with code of %v and err of: %v\n", code, err)
}

// Case commands for Mdserv
func ls(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get current dir
	// NOTE: here and in other locations, trimming whitespace may be more desirable
	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	// get the length of arguments to the ls command
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// the message which will be returned to the user, each entry will be
	// comma separated for the client to interpret
	msg := ""

	// if only the ls command was called
	if lenArgs == 1 {

		fmt.Println(md.getPath() + "files" + currentDir)

		files, err := ioutil.ReadDir(md.getPath() + "files" + currentDir)
		if err != nil {
			w.Flush()
		}

		// iterate over the files, and comma separate them while appending to msg
		for _, file := range files {
			if !utils.IsHidden(file.Name()) {
				msg = msg + file.Name() + ","
			}
		}

	}

	// loop for dealing with one or more args
	for i := 1; i < int(lenArgs); i++ {

		// reading in this arg
		targetPath, _ := r.ReadString('\n')
		targetPath = strings.TrimSpace(targetPath)

		if !path.IsAbs(targetPath) {
			targetPath = path.Join(currentDir, targetPath)
		}

		fmt.Printf("  in loop read in targetPath: %s\n", (targetPath))
		files, err := ioutil.ReadDir(md.getPath() + "files" + targetPath)
		if err != nil {

			// if it is not a directory, skip it and try the next arg
			continue

		} else {

			msg = msg + targetPath + ":," // note comma to denote newline
			for _, file := range files {
				if !utils.IsHidden(file.Name()) {
					msg = msg + file.Name() + ","
				}
			}
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
	return nil
}

func mkdir(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get currentDir
	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	// get lenArgs for mkdir
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// if no more than "mkdir" is sent, nothing will happen
	// if errors occur, the dir will just not be made
	for i := 1; i < int(lenArgs); i++ {

		fmt.Printf("  in loop at pos %d ready to read\n", i)

		// for each arg, get the target path
		targetPath, _ := r.ReadString('\n')
		targetPath = strings.TrimSpace(targetPath)

		if !path.IsAbs(targetPath) {
			targetPath = path.Join(currentDir, targetPath)
		}

		// print the target for terminal's sake
		fmt.Printf("  in loop read in targetPath: %s", targetPath)

		// MkdirAll creates an entire file path if some dirs are missing
		if !utils.IsHidden(targetPath) {
			os.MkdirAll(md.getPath()+"files"+targetPath, 0777)
		}
	}
	return nil
}

func rmdir(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get currentDir
	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	// get len args for rmdir
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// only does something if more than "rmdir" is called, as above
	for i := 1; i < int(lenArgs); i++ {

		fmt.Printf("  in loop at pos %d ready to read\n", i)

		targetPath, _ := r.ReadString('\n')
		targetPath = strings.TrimSpace(targetPath)

		if !path.IsAbs(targetPath) {
			targetPath = path.Join(currentDir, targetPath)
		}

		fmt.Printf("  in loop read in targetPath: %s", targetPath)

		// this will only remove a dir that is empty, else it does nothing
		// BUG-NOTE: this command will also currently delete files (there is not
		// a different command to rmdir an rm in golang), so a check to make sure
		// the targetPath is a dir should take place (sample code for checking if
		// a path is a dir or a file is found in "cd" below).
		// NOTE: a nice to have would be a recursive remove similar to rm -rf,
		// but this is not needed

		src, err := os.Stat(md.getPath() + "files" + targetPath)
		if !utils.IsHidden(targetPath) && err == nil && src.IsDir() {
			os.Remove(md.getPath() + "files" + targetPath)
		}
	}
	return nil
}

func rm(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get currentDir
	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	// get len args for rmdir
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// only does something if more than "rmdir" is called, as above
	for i := 1; i < int(lenArgs); i++ {

		fmt.Printf("  in loop at pos %d ready to read\n", i)

		targetPath, _ := r.ReadString('\n')
		targetPath = strings.TrimSpace(targetPath)

		if !path.IsAbs(targetPath) {
			targetPath = path.Join(currentDir, targetPath)
		}

		fmt.Printf("  in loop read in targetPath: %s", targetPath)

		// this will only remove a dir that is empty, else it does nothing
		// BUG-NOTE: this command will also currently delete files (there is not
		// a different command to rmdir an rm in golang), so a check to make sure
		// the targetPath is a dir should take place (sample code for checking if
		// a path is a dir or a file is found in "cd" below).
		// NOTE: a nice to have would be a recursive remove similar to rm -rf,
		// but this is not needed

		src, err := os.Stat(md.getPath() + "files" + targetPath)
		if !utils.IsHidden(targetPath) && err == nil && !src.IsDir() {
			os.Remove(md.getPath() + "files" + targetPath)
		}
	}
	return nil
}

func cd(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get current dir and target path
	currentDir, _ := r.ReadString('\n')
	targetPath, _ := r.ReadString('\n')

	currentDir = strings.TrimSpace(currentDir)
	targetPath = strings.TrimSpace(targetPath)

	if !path.IsAbs(targetPath) {
		targetPath = path.Join(currentDir, targetPath)
	}

	// print for terminal's sake
	fmt.Printf("currentDir = %s\n", currentDir)
	fmt.Printf("targetPath = %s\n", targetPath)

	// check if the source dir exist
	src, err := os.Stat(md.getPath() + "files" + targetPath)
	if err != nil || utils.IsHidden(targetPath) { // not a path ie. not a dir OR a file

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
			fmt.Printf("Path \"%s\" is a directory\n", targetPath)

			// send the new path back to the user
			w.WriteString(targetPath + "\n")
			w.Flush()
		}
	}
	return nil
}

func request(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	//get currentDir
	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	// receive filename from client
	filename, _ := r.ReadString('\n')
	filename = strings.TrimSpace(filename)

	// Cleans the filepath for proper access use

	if !path.IsAbs(filename) {
		filename = path.Join(currentDir, filename)
	}

	// check if the filename exists
	src, err := os.Stat(md.getPath() + "files" + filename)
	if err != nil || utils.IsHidden(filename) { // not a path ie. not a dir OR a file

		fmt.Println("File \"" + filename + "\" does not exist")

		// notify the client that it is not existant with code "2"
		w.WriteByte(1)
		w.Flush()
		return nil

	} else if src.IsDir() { // notify the client that the file exists with code "1"

		fmt.Println("Path \"" + filename + "\" is a directory")

		// notify that is a dir
		w.WriteByte(2)
		w.Flush()
	} else {

		fmt.Println("File \"" + filename + "\" exists")

		// notify success
		w.WriteByte(3)
		w.Flush()
	}

	hash, unid, protected, err := getFile(md.getPath() + "files" + filename)

	if protected {
		w.WriteByte(1)
		w.Flush()
	} else {
		w.WriteByte(2)
		w.Flush()
	}

	fmt.Println(hash + ", " + unid)

	md.stnodeDB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		fmt.Println(2)
		b := tx.Bucket([]byte("stnodes"))

		v := b.Get([]byte(unid))
		fmt.Println(3)

		if v == nil {

			fmt.Println("No stnode for: " + unid)
			w.WriteByte(1)
			w.Flush()
			return nil
		}

		w.WriteByte(2)
		w.Flush()

		var tmpStnode utils.Stnode
		json.Unmarshal(v, &tmpStnode)

		fmt.Println(tmpStnode)

		fmt.Println("protocol: " + tmpStnode.Protocol + ", address: " + tmpStnode.NAddress)

		w.WriteString(hash + "\n")
		w.WriteString(tmpStnode.Protocol + "\n")
		w.WriteString(tmpStnode.NAddress + "\n")
		w.Flush()

		return nil
	})

	return nil
}

func send(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get current dir
	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	// receive filename from client
	filename, _ := r.ReadString('\n')
	filename = strings.TrimSpace(filename)

	// clean the path
	if !path.IsAbs(filename) {
		filename = path.Join(currentDir, filename)
	}

	// check if the filename exists already
	_, err = os.Stat(md.getPath() + "files" + filename)
	if err != nil && !utils.IsHidden(filename) { // not a path ie. not a dir OR a file

		fmt.Println("File \"" + filename + "\" does not exist")

		// notify the client that it is not already on system
		w.WriteByte(2)
		w.Flush()

	} else { // notify the client that the file exists with error code "1"

		w.WriteByte(1)
		w.Flush()

		return nil
	}

	// is the user encrypting the file?
	protected := false
	enc, _ := r.ReadByte()
	if enc == 1 { // we are encrypting

		protected = true
		fmt.Println("Receiving encrypted file")
		// send the pubkeys of users here

		// get lenArgs
		lenArgs, _ := r.ReadByte()

		for i := 0; i < int(lenArgs); i++ {
			fmt.Printf("About to query DB for time %d of %d\n", i, int(lenArgs))
			md.userDB.View(func(tx *bolt.Tx) error {

				b := tx.Bucket([]byte("users"))

				// get the proposed uuid for a user
				uuid, _ := r.ReadString('\n')
				uuid = strings.TrimSpace(uuid)
				uuidUint64, _ := strconv.ParseUint(uuid, 10, 64)

				v := b.Get(itob(uuidUint64))

				if v == nil {

					fmt.Println("No user profile matching uuid: " + uuid)
					w.WriteString("INV" + "\n")
					w.Flush()
					return nil
				}

				var tmpUser utils.User
				json.Unmarshal(v, &tmpUser)

				fmt.Println("Found user: " + tmpUser.Uname)

				w.WriteString(uuid + "\n")
				w.Write([]byte(tmpUser.Pubkey.N.String() + "\n"))
				w.Write([]byte(strconv.Itoa(tmpUser.Pubkey.E) + "\n"))
				w.Flush()

				return nil
			})

		}

	} else if enc == 2 {

		//invalid cmd on client side
		return nil
		// else we are not
	}

	// get the hash of the file
	hash := utils.ReadHashAsString(r)

	var success byte
	var unid string

	fmt.Println(1)
	md.stnodeDB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("stnodes"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			fmt.Println(k)
			fmt.Println(v)

			w.WriteByte(1) // got a stnode
			w.Flush()

			var tmpStnode utils.Stnode
			json.Unmarshal(v, &tmpStnode)

			fmt.Println(tmpStnode)

			fmt.Println("protocol: " + tmpStnode.Protocol + ", address: " + tmpStnode.NAddress)

			w.WriteString(tmpStnode.Protocol + "\n")
			w.WriteString(tmpStnode.NAddress + "\n")
			unid = tmpStnode.Unid
			w.Flush()

			success, _ = r.ReadByte()
			if success != 1 {
				fmt.Println("Successful send to stnode from client")
				return nil
			}
		}

		w.WriteByte(2) // no more stnodes
		w.Flush()

		return nil
	})

	if success != 2 {
		fmt.Println("No stnodes were available to the client")
		return nil
	}

	success, _ = r.ReadByte()
	if success != 1 {
		fmt.Println("Error on client side sending file to stnode")
		return nil
	}

	err = createFile(md.getPath()+"files"+filename, hash, unid, protected)
	if err != nil {
		panic(err)
	}
	return nil
}

func newUser(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get the uuid for the new user
	var newUser utils.User
	err = md.userDB.Update(func(tx *bolt.Tx) (err error) {

		// Retrieve the users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("users"))

		// Generate ID for the user.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, _ := b.NextSequence()
		newUser.Uuid = uint64(id)
		idStr := strconv.FormatUint(id, 10)

		// receive the username	and public key for the new user
		uname, _ := r.ReadString('\n')
		pubKN, _ := r.ReadString('\n')
		pubKE, _ := r.ReadString('\n')

		newUser.Uname = strings.TrimSpace(uname)
		newUser.Pubkey = &rsa.PublicKey{N: big.NewInt(0)}
		newUser.Pubkey.N.SetString(strings.TrimSpace(pubKN), 10)

		newUser.Pubkey.E, err = strconv.Atoi(strings.TrimSpace(pubKE))

		fmt.Println("New user: " + newUser.Uname)
		fmt.Println("recieved key")
		fmt.Println("key stored in new user")

		// Marshal user data into bytes.
		buf, err := json.Marshal(newUser)
		if err != nil {
			return err
		}

		fmt.Println("writing uuid")
		w.WriteString(idStr + "\n")
		fmt.Println("written")

		w.Flush()
		fmt.Println("flushed")

		// Persist bytes to users bucket.
		return b.Put(itob(newUser.Uuid), buf)
	})

	return err
}

func newStnode(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get unid for a new stnode
	var newStnode utils.Stnode
	err = md.stnodeDB.Update(func(tx *bolt.Tx) (err error) {

		// Retrieve the users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("stnodes"))

		// Generate ID for the stnode.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, _ := b.NextSequence()
		idStr := strconv.FormatUint(id, 10)
		newStnode.Unid = idStr

		// Receive the connection type and the address to be used for
		// conneting to the stnode
		protocol, _ := r.ReadString('\n')
		nAddress, _ := r.ReadString('\n')

		newStnode.Protocol = strings.TrimSpace(protocol)
		newStnode.NAddress = strings.TrimSpace(nAddress)

		fmt.Println("Received stnode " + idStr + "'s protocol: " + newStnode.Protocol)
		fmt.Println("Received stnode " + idStr + "'s network address: " + newStnode.NAddress)

		// Marshal stnode data into bytes.
		buf, err := json.Marshal(newStnode)
		if err != nil {
			return err
		}

		fmt.Println("writing unid to stnode")

		w.WriteString(idStr + "\n")
		fmt.Println("written")

		w.Flush()
		fmt.Println("flushed")

		// Persist bytes to stnodes bucket.
		return b.Put([]byte(idStr), buf)
	})

	return err
}
