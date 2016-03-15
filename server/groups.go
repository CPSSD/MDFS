package server

import (
	"bufio"
	//"crypto/rsa"
	//"encoding/binary"
	"encoding/json"
	"fmt"
	//"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"github.com/boltdb/bolt"
	//"io/ioutil"
	//"math/big"
	"net"
	//"os"
	//"path"
	"strconv"
	"strings"
)

func createGroup(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	fmt.Println("Called create group")

	// get lenArgs
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// get details for owner of new group
	uuid, _ := r.ReadString('\n')
	uintUuid, err := strconv.ParseUint(strings.TrimSpace(uuid), 10, 64)
	if err != nil {
		return err
	}

	for i := 1; i < int(lenArgs); i++ {

		fmt.Printf("  in loop at pos %d ready to read\n", i)

		var newGroup utils.Group

		groupName, _ := r.ReadString('\n')
		groupName = strings.TrimSpace(groupName)
		fmt.Printf("  in loop read in groupName: %s", groupName)

		// create the group in the database
		err = md.userDB.Update(func(tx *bolt.Tx) (err error) {

			// get group bucket
			b := tx.Bucket([]byte("groups"))

			id, _ := b.NextSequence()
			idStr := strconv.FormatUint(id, 10)

			newGroup.Gid = uint64(id)
			newGroup.Gname = groupName
			newGroup.Members = append(newGroup.Members, uintUuid)
			newGroup.Owner = uintUuid

			fmt.Println("New group \"" + newGroup.Gname + "\" with owner id: " + strings.TrimSpace(uuid))

			buf, err := json.Marshal(newGroup)
			if err != nil {
				return err
			}

			fmt.Println("writing Gid")
			w.WriteString(idStr + "\n")
			fmt.Println("written")
			w.Flush()

			return b.Put(itob(newGroup.Gid), buf)
		})
		if err != nil {
			return err
		}
	}

	return err
}
