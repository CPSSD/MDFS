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

func groupAdd(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get lenArgs
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// get details for current accessor
	uuid, _ := r.ReadString('\n')
	uintUuid, err := strconv.ParseUint(strings.TrimSpace(uuid), 10, 64)
	if err != nil {
		return err
	}

	// get details for group to add to
	gid, _ := r.ReadString('\n')
	gid = strings.TrimSpace(gid)
	uintGid, err := strconv.ParseUint(strings.TrimSpace(gid), 10, 64)
	if err != nil {
		return err
	}

	err = md.userDB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("groups"))

		v := b.Get(itob(uintGid))

		fmt.Println("Trying to get: " + gid)

		if v == nil {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("Bad access")
		}

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		if tmpGroup.Owner != uintUuid {

			fmt.Printf("Owner: %d, and uuid: %d", tmpGroup.Owner, uintUuid)

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("Bad access")
		}

		// authorised and group exists
		w.WriteByte(1)
		w.Flush()

		return nil
	})

	if err != nil {

		fmt.Println("Invalid access to group: " + strings.TrimSpace(gid))
		return nil
	}

	var users []uint64

	// read in all users
	for i := 2; i < int(lenArgs); i++ {

		fmt.Printf("  in loop at pos %d ready to read\n", i)

		user, _ := r.ReadString('\n')
		user = strings.TrimSpace(user)

		fmt.Printf("  in loop read in user to add: %s\n", user)

		exists, _ := userExists(user, md.userDB)
		if exists {

			uintUuid, _ := strconv.ParseUint(strings.TrimSpace(user), 10, 64)
			users = append(users, uintUuid)
		}
	}

	// add users to the group in the database
	err = md.userDB.Update(func(tx *bolt.Tx) (err error) {

		// get group bucket
		b := tx.Bucket([]byte("groups"))

		// we already know it exists from above
		v := b.Get([]byte(strings.TrimSpace(gid)))

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		newUsers := ""

		for _, u := range users {
			if !utils.Contains(u, tmpGroup.Members) {
				tmpGroup.Members = append(tmpGroup.Members, u)
				fmt.Println(u)
				newUsers = newUsers + strconv.FormatUint(u, 10) + ", "
			}
		}

		buf, err := json.Marshal(tmpGroup)
		if err != nil {
			return err
		}

		fmt.Println("writing users added")
		w.WriteString(newUsers + "\n")
		fmt.Println("written")
		w.Flush()

		return b.Put(itob(tmpGroup.Gid), buf)
	})

	if err != nil {
		return err
	}
	return err
}

func groupRemove(conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get lenArgs
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// get details for current accessor
	uuid, _ := r.ReadString('\n')
	uintUuid, err := strconv.ParseUint(strings.TrimSpace(uuid), 10, 64)
	if err != nil {
		return err
	}

	// get details for group to remove from
	gid, _ := r.ReadString('\n')
	gid = strings.TrimSpace(gid)
	uintGid, err := strconv.ParseUint(strings.TrimSpace(gid), 10, 64)
	if err != nil {
		return err
	}

	// ensure valid user and group
	err = md.userDB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("groups"))

		v := b.Get(itob(uintGid))

		fmt.Println("Trying to get: " + gid)

		if v == nil {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("Bad access")
		}

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		if tmpGroup.Owner != uintUuid {

			fmt.Printf("Owner: %d, and uuid: %d", tmpGroup.Owner, uintUuid)

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("Bad access")
		}

		// authorised and group exists
		w.WriteByte(1)
		w.Flush()

		return nil
	})

	if err != nil {

		fmt.Println("Invalid access to group: " + strings.TrimSpace(gid))
		return nil
	}

	var users []uint64

	// read in all users
	for i := 2; i < int(lenArgs); i++ {

		fmt.Printf("  in loop at pos %d ready to read\n", i)

		user, _ := r.ReadString('\n')
		user = strings.TrimSpace(user)

		fmt.Printf("  in loop read in user to remove: %s\n", user)

		exists, _ := userExists(user, md.userDB)
		if exists {

			uintUuid, _ := strconv.ParseUint(strings.TrimSpace(user), 10, 64)
			users = append(users, uintUuid)
		}
	}

	// add users to the group in the database
	err = md.userDB.Update(func(tx *bolt.Tx) (err error) {

		// get group bucket
		b := tx.Bucket([]byte("groups"))

		// we already know it exists from above
		v := b.Get([]byte(strings.TrimSpace(gid)))

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		removedUsers := ""

		for _, u := range users {
			for i, member := range tmpGroup.Members {
				if member == u {

					fmt.Println(u)
					tmpGroup.Members = append(tmpGroup.Members[:i], tmpGroup.Members[i+1:]...)
					removedUsers = removedUsers + strconv.FormatUint(u, 10) + ", "
					break
				}
			}
		}

		buf, err := json.Marshal(tmpGroup)
		if err != nil {
			return err
		}

		fmt.Println("writing users removed")
		w.WriteString(removedUsers + "\n")
		fmt.Println("written")
		w.Flush()

		return b.Put(itob(tmpGroup.Gid), buf)
	})

	if err != nil {
		return err
	}
	return err
}

func userExists(uuid string, db *bolt.DB) (exists bool, err error) {

	exists = false
	err = db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("users"))
		v := b.Get([]byte(uuid))
		if v != nil {
			exists = true
		}
		return nil
	})
	return
}
