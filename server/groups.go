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
	"path"
	"strconv"
	"strings"
)

func createGroup(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	fmt.Println("Called create group")

	// get lenArgs
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

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
			newGroup.Members = append(newGroup.Members, uuid)
			newGroup.Owner = uuid

			fmt.Println("New group \"" + newGroup.Gname + "\" with owner id: " + strconv.FormatUint(uuid, 10))

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

func groupAdd(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get lenArgs
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// get details for group to add to
	gid, _ := r.ReadString('\n')
	gid = strings.TrimSpace(gid)
	uintGid, err := strconv.ParseUint(strings.TrimSpace(gid), 10, 64)
	if err != nil {
		fmt.Println("Not a uint or gid")
		w.WriteByte(2)
		w.Flush()
		return nil
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

		if tmpGroup.Owner != uuid {

			fmt.Printf("Owner: %d, and uuid: %d", tmpGroup.Owner, uuid)

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

			fmt.Println("Exists")
			uintUuid, _ := strconv.ParseUint(strings.TrimSpace(user), 10, 64)
			users = append(users, uintUuid)
		}
	}

	// add users to the group in the database
	err = md.userDB.Update(func(tx *bolt.Tx) (err error) {

		// get group bucket
		b := tx.Bucket([]byte("groups"))

		// we already know it exists from above
		v := b.Get(itob(uintGid))

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		newUsers := ""

		for _, u := range users {
			fmt.Printf("Checking user %d if exists in \n", u)

			if !utils.Contains(u, tmpGroup.Members) {
				tmpGroup.Members = append(tmpGroup.Members, u)
				fmt.Printf("Adding user: %d\n", u)
				newUsers = newUsers + strconv.FormatUint(u, 10) + ", "
			}
		}

		buf, err := json.Marshal(tmpGroup)
		if err != nil {
			return err
		}

		var anGroup utils.Group
		json.Unmarshal(buf, &anGroup)

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

func groupRemove(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get lenArgs
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	// get details for group to remove from
	gid, _ := r.ReadString('\n')
	gid = strings.TrimSpace(gid)
	uintGid, err := strconv.ParseUint(strings.TrimSpace(gid), 10, 64)
	if err != nil {
		fmt.Println("Not a uint or gid")
		w.WriteByte(2)
		w.Flush()
		return nil
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

		if tmpGroup.Owner != uuid {

			fmt.Printf("Owner: %d, and uuid: %d", tmpGroup.Owner, uuid)

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
		v := b.Get(itob(uintGid))

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

func groupLs(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get details for group to list
	gid, _ := r.ReadString('\n')
	gid = strings.TrimSpace(gid)
	uintGid, err := strconv.ParseUint(gid, 10, 64)
	if err != nil {
		fmt.Println("Not a uint or gid")
		w.WriteByte(2)
		w.Flush()
		return nil
	}

	var members []uint64

	// ensure valid group
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

		members = tmpGroup.Members

		// group exists
		w.WriteByte(1)
		w.Flush()

		return nil
	})

	if err != nil {

		fmt.Println("Invalid access to group: " + gid)
		return nil
	}

	result := ""

	for i, member := range members {

		fmt.Printf("Member %d = uuid of %d\n", i, member)
		result = result + strconv.FormatUint(member, 10) + ", "
	}

	fmt.Println("writing members of group")
	w.WriteString(result + "\n")
	fmt.Println("written")
	w.Flush()

	if err != nil {
		return err
	}
	return err
}

func deleteGroup(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	// get details for group to remove from
	gid, _ := r.ReadString('\n')
	gid = strings.TrimSpace(gid)
	uintGid, err := strconv.ParseUint(strings.TrimSpace(gid), 10, 64)
	if err != nil {
		fmt.Println("Not a uint or gid")
		w.WriteByte(2)
		w.Flush()
		return nil
	}

	// add users to the group in the database
	md.userDB.Update(func(tx *bolt.Tx) (err error) {

		b := tx.Bucket([]byte("groups"))

		v := b.Get(itob(uintGid))

		fmt.Println("Trying to get: " + gid)

		if v == nil {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("Bad access: not a group")
		}

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		if tmpGroup.Owner != uuid {

			fmt.Printf("Owner: %d, and uuid: %d\n", tmpGroup.Owner, uuid)

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("Bad access: not owner")
		}

		// authorised and group exists
		w.WriteByte(1)
		w.Flush()

		return b.Delete(itob(tmpGroup.Gid))
	})
	return nil
}

func userExists(uuid string, db *bolt.DB) (exists bool, err error) {

	exists = false
	uuidUint, err := strconv.ParseUint(strings.TrimSpace(uuid), 10, 64)
	if err != nil {
		return false, nil
	}

	err = db.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("users"))
		v := b.Get(itob(uuidUint))
		if v != nil {
			fmt.Println("v != nil")
			exists = true
		}
		fmt.Println("v == nil")

		return nil
	})
	return
}

func listGroups(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	mod, _ := r.ReadByte()

	fmt.Println(mod)

	switch mod {
	case 1: // sendcode 26
		return listGroupsMemberOf(uuid, conn, r, w, md)
	case 2: // sendcode 27
		return listGroupsOwnerOf(uuid, conn, r, w, md)
	case 3:
		return listGroupsAll(uuid, conn, r, w, md)
	case 4:
		return nil
	}
	return nil
}

func listGroupsAll(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	fmt.Printf("User listing: %d\n", uuid)

	err = md.userDB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("groups"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			var tmpGroup utils.Group
			json.Unmarshal(v, &tmpGroup)

			w.WriteByte(1)
			w.Flush()

			groups := ""

			groups = groups + tmpGroup.Gname + "," + strconv.FormatUint(tmpGroup.Gid, 10) + ","
			for _, u := range tmpGroup.Members {
				groups = groups + strconv.FormatUint(u, 10) + ","
			}

			w.WriteString(groups + "\n")
			w.Flush()

			fmt.Println("Written group " + tmpGroup.Gname)
		}

		w.WriteByte(2)
		w.Flush() // indicate to client that we are done listing

		return nil
	})

	return
}

func listGroupsOwnerOf(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	err = md.userDB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("groups"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			var tmpGroup utils.Group
			json.Unmarshal(v, &tmpGroup)

			if tmpGroup.Owner != uuid {

				continue
			}

			w.WriteByte(1)
			w.Flush()

			groups := ""

			groups = groups + tmpGroup.Gname + "," + strconv.FormatUint(tmpGroup.Gid, 10) + ","
			for _, u := range tmpGroup.Members {
				groups = groups + strconv.FormatUint(u, 10) + ","
			}

			w.WriteString(groups + "\n")
			w.Flush()

			fmt.Println("Written group " + tmpGroup.Gname)
		}

		w.WriteByte(2)
		w.Flush() // indicate to client that we are done listing

		return nil
	})

	return
}

func listGroupsMemberOf(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	err = md.userDB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("groups"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			var tmpGroup utils.Group
			json.Unmarshal(v, &tmpGroup)

			if !utils.Contains(uuid, tmpGroup.Members) {

				continue
			}

			w.WriteByte(1)
			w.Flush()

			groups := ""

			groups = groups + tmpGroup.Gname + "," + strconv.FormatUint(tmpGroup.Gid, 10) + ","
			for _, u := range tmpGroup.Members {
				groups = groups + strconv.FormatUint(u, 10) + ","
			}

			w.WriteString(groups + "\n")
			w.Flush()

			fmt.Println("Written group " + tmpGroup.Gname)
		}

		w.WriteByte(2)
		w.Flush() // indicate to client that we are done listing

		return nil
	})

	return
}

func checkBase(uuid uint64, targetPath string, mod string, md *MDService) (auth bool) {

	basePath := strings.TrimSuffix(targetPath, "/"+path.Base(targetPath))
	fmt.Println("Checking basePath: " + basePath)
	return checkEntry(uuid, basePath, mod, md)
}

func checkEntry(uuid uint64, targetPath, mod string, md *MDService) (auth bool) {

	// check all the d in dirs for Xecute
	dirs := strings.Split(targetPath, "/")
	if targetPath == "/" || targetPath == "" {
		fmt.Println("Root dir")
		return true
	}

	traverser := ""

	for i, d := range dirs {
		if i != 0 {
			traverser = path.Join("/", traverser, d)
			owner, groups, permissions, err := getPerm(md.getPath() + "files/" + traverser + "/")
			if err != nil {
				fmt.Println("NO PERM FILE AT: " + md.getPath() + "files" + traverser + "/.perm")
				return false
			}

			fmt.Printf("%d, %s, %d\n", i, d, owner)

			hasGroup := false
			if owner == uuid {

				return true

			} else if groups != nil {

				for _, g := range groups {
					err = md.userDB.View(func(tx *bolt.Tx) error {

						b := tx.Bucket([]byte("groups"))

						v := b.Get(itob(g))

						if v == nil {
							return nil
						}

						var tmpGroup utils.Group
						json.Unmarshal(v, &tmpGroup)

						if utils.Contains(uuid, tmpGroup.Members) {
							hasGroup = true
						}
						return nil
					})
					if hasGroup {
						break
					}
				}
			} else {

				switch mod {
				case "r":
					return (hasGroup && permissions[0]) || permissions[3]

				case "w":
					return (hasGroup && permissions[1]) || permissions[4]

				case "x":
					return (hasGroup && permissions[2]) || permissions[5]
				}
			}
		}
	}

	return false
}
