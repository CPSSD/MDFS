package server

import (
	"bufio"
	//"crypto/rsa"
	//"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
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
		fmt.Printf("  in loop read in groupName: %s\n", groupName)

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

			fmt.Printf("Owner: %d, and uuid: %d\n", tmpGroup.Owner, uuid)

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

			fmt.Printf("Owner: %d, and uuid: %d\n", tmpGroup.Owner, uuid)

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

func checkFile(uuid uint64, targetPath, mod string, md *MDService) (auth bool) {

	fmt.Println("Getting filestats for: " + path.Join(md.getPath(), "files", targetPath))
	_, _, _, owner, groups, permissions, err := getFile(path.Join(md.getPath(), "files", targetPath))
	if err != nil {
		fmt.Println("NO FILE AT: " + md.getPath() + "files" + targetPath)
		return false
	}

	fmt.Printf("filestats: %d, %v, %v\n", owner, groups, permissions)

	hasGroup := false
	if groups != nil {

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

	}

	auth = (owner == uuid) || (hasGroup && permissions[0]) || permissions[1]

	base := checkBase(uuid, targetPath, mod, md)

	fmt.Printf("%b == %b, %b\n", auth, base, permissions[1])

	return base && auth
}

func checkBase(uuid uint64, targetPath, mod string, md *MDService) (auth bool) {

	basePath := strings.TrimSuffix(targetPath, "/"+path.Base(targetPath))
	fmt.Println("Checking basePath: " + basePath)
	return checkEntry(uuid, basePath, mod, md)
}

func checkEntry(uuid uint64, targetPath, mod string, md *MDService) (auth bool) {

	// check all the d in dirs for Xecute
	dirs := strings.Split(targetPath, "/")
	if targetPath == "/" || targetPath == "" {
		fmt.Println("Root dir access")
		return true
	}

	traverser := ""
	auth = true

	for i, d := range dirs {

		traverser = path.Join("/", traverser, d)
		fmt.Printf("Auth = %b, Traverser = %s\n", auth, traverser)
		if i != 0 {

			owner, groups, permissions, err := getPerm(md.getPath() + "files/" + traverser + "/")
			if err != nil {
				fmt.Println("NO PERM FILE AT: " + md.getPath() + "files" + traverser + "/.perm")
				return false
			}

			fmt.Printf("%d, %s, %d\n", i, d, owner)

			hasGroup := false
			if groups != nil {

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
			}
			auth = (owner == uuid) || (hasGroup && permissions[2]) || permissions[5]
		}
		if !auth {
			return auth
		}

		if i == len(dirs)-1 {
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
			}

			switch mod {
			case "r":
				return (hasGroup && permissions[0]) || permissions[3]

			case "w":
				fmt.Printf("Checking w, perm = %b\n", permissions[4])
				return (hasGroup && permissions[1]) || permissions[4]

			case "x":
				return (hasGroup && permissions[2]) || permissions[5]
			}
		}
	}

	fmt.Printf("Auth = %b, Traverser = %s\n", auth, traverser)
	return auth
}

func permit(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	flag, _ := r.ReadString('\n')
	flag = strings.TrimSpace(flag)
	if flag == "INV" {
		fmt.Println("Invalid flag from client")
		return nil
	}
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	targetPath, _ := r.ReadString('\n')
	targetPath = strings.TrimSpace(targetPath)

	if !path.IsAbs(targetPath) {
		targetPath = path.Join(currentDir, targetPath)
	}

	fmt.Println("Target = " + md.getPath() + "files" + targetPath)

	src, err := os.Stat(md.getPath() + "files" + targetPath)
	if !utils.IsHidden(targetPath) && err == nil {
		// exists, not hidden path
	}

	var groups []uint64

	if src.IsDir() {
		addPerms, _ := r.ReadString('\n')
		addPerms = strings.TrimSpace(addPerms)
		fmt.Println("HERE IN PERMIT")

		for i := 4; i < int(lenArgs); i++ {
			group, _ := r.ReadString('\n')
			gid, err := strconv.ParseUint(strings.TrimSpace(group), 10, 64)
			if err != nil {
				continue
			}
			groups = append(groups, gid)
		}
		fmt.Println("HERE IN PERMIT")

		if checkEntry(uuid, targetPath, "owner", md) {

			owner, existingGroups, permissions, err := getPerm(md.getPath() + "files" + targetPath)
			if err != nil {
				fmt.Println("Error finding .perm for dir: " + targetPath)
				return nil
			}

			switch flag {
			case "-g":
				for _, g := range groups {

					if !utils.Contains(g, existingGroups) {
						existingGroups = append(existingGroups, g)
						fmt.Printf("Permitting group: %d\n", g)
					}
				}
				if strings.Contains(addPerms, "r") {
					permissions[0] = true

				}
				if strings.Contains(addPerms, "w") {
					permissions[1] = true

				}
				if strings.Contains(addPerms, "x") {
					permissions[2] = true

				}

			case "-w":
				if strings.Contains(addPerms, "r") {
					permissions[3] = true

				}
				if strings.Contains(addPerms, "w") {
					permissions[4] = true

				}
				if strings.Contains(addPerms, "x") {
					permissions[5] = true

				}
			}

			err = createPerm(md.getPath()+"files"+targetPath, owner, existingGroups, permissions)
			if err != nil {
				fmt.Println("Error re-writing .perm for dir: " + targetPath)
			}
		}

	} else {

		for i := 3; i < int(lenArgs); i++ {
			group, _ := r.ReadString('\n')
			gid, err := strconv.ParseUint(strings.TrimSpace(group), 10, 64)
			if err != nil {
				continue
			}
			groups = append(groups, gid)
		}

		if checkFile(uuid, targetPath, "x", md) {
			hash, stnode, protected, owner, existingGroups, permissions, err := getFile(md.getPath() + "files" + targetPath)
			if err != nil {
				fmt.Println("Error finding file_perms for file: " + targetPath)
				return nil
			}

			switch flag {
			case "-g":
				for _, g := range groups {

					if !utils.Contains(g, existingGroups) {
						existingGroups = append(existingGroups, g)
						fmt.Printf("Permitting group: %d\n", g)
					}
				}
				permissions[0] = true

			case "-w":
				permissions[1] = true
			}

			err = createFile(md.getPath()+"files"+targetPath, hash, stnode, protected, owner, existingGroups, permissions)
			if err != nil {
				fmt.Println("Error re-writing .perm for dir: " + targetPath)
			}
		}
	}
	return nil
}

func deny(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	flag, _ := r.ReadString('\n')
	flag = strings.TrimSpace(flag)
	if flag == "INV" {
		fmt.Println("Invalid flag from client")
		return nil
	}
	lenArgs, _ := r.ReadByte()
	fmt.Printf("lenArgs = %v\n", lenArgs)

	targetPath, _ := r.ReadString('\n')
	targetPath = strings.TrimSpace(targetPath)

	if !path.IsAbs(targetPath) {
		targetPath = path.Join(currentDir, targetPath)
	}

	fmt.Println("Target = " + md.getPath() + "files" + targetPath)

	src, err := os.Stat(md.getPath() + "files" + targetPath)
	if !utils.IsHidden(targetPath) && err == nil {
		// exists, not hidden path
	}

	var groups []uint64

	if src.IsDir() {
		addPerms, _ := r.ReadString('\n')
		addPerms = strings.TrimSpace(addPerms)
		fmt.Println("HERE IN PERMIT")

		for i := 4; i < int(lenArgs); i++ {
			group, _ := r.ReadString('\n')
			gid, err := strconv.ParseUint(strings.TrimSpace(group), 10, 64)
			if err != nil {
				continue
			}
			groups = append(groups, gid)
		}
		fmt.Println("HERE IN PERMIT")

		if checkEntry(uuid, targetPath, "owner", md) {

			owner, existingGroups, permissions, err := getPerm(md.getPath() + "files" + targetPath)
			if err != nil {
				fmt.Println("Error finding .perm for dir: " + targetPath)
				return nil
			}

			switch flag {
			case "-g":
				for _, g := range groups {
					for i, gr := range existingGroups {
						if g == gr {
							existingGroups = append(existingGroups[:i], existingGroups[i+1:]...)
							fmt.Printf("Denying group: %d\n", g)
							break
						}
					}
				}
				if strings.Contains(addPerms, "r") {
					permissions[0] = false

				}
				if strings.Contains(addPerms, "w") {
					permissions[1] = false

				}
				if strings.Contains(addPerms, "x") {
					permissions[2] = false

				}

			case "-w":
				if strings.Contains(addPerms, "r") {
					permissions[3] = false

				}
				if strings.Contains(addPerms, "w") {
					permissions[4] = false

				}
				if strings.Contains(addPerms, "x") {
					permissions[5] = false

				}
			}

			err = createPerm(md.getPath()+"files"+targetPath, owner, existingGroups, permissions)
			if err != nil {
				fmt.Println("Error re-writing .perm for dir: " + targetPath)
			}
		}

	} else {

		for i := 3; i < int(lenArgs); i++ {
			group, _ := r.ReadString('\n')
			gid, err := strconv.ParseUint(strings.TrimSpace(group), 10, 64)
			if err != nil {
				continue
			}
			groups = append(groups, gid)
		}

		if checkFile(uuid, targetPath, "x", md) {
			hash, stnode, protected, owner, existingGroups, permissions, err := getFile(md.getPath() + "files" + targetPath)
			if err != nil {
				fmt.Println("Error finding file_perms for file: " + targetPath)
				return nil
			}

			switch flag {
			case "-g":
				for _, g := range groups {
					for i, gr := range existingGroups {
						if g == gr {
							existingGroups = append(existingGroups[:i], existingGroups[i+1:]...)
							fmt.Printf("Denying group: %d\n", g)
							break
						}
					}
				}
				permissions[0] = false

			case "-w":
				permissions[1] = false
			}

			err = createFile(md.getPath()+"files"+targetPath, hash, stnode, protected, owner, existingGroups, permissions)
			if err != nil {
				fmt.Println("Error re-writing .perm for dir: " + targetPath)
			}
		}
	}
	return nil
}
