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

	// get lenArgs
	lenArgs, _ := r.ReadByte()

	for i := 1; i < int(lenArgs); i++ {

		var newGroup utils.Group

		groupName, _ := r.ReadString('\n')
		groupName = strings.TrimSpace(groupName)

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

			fmt.Println("\tNew group \"" + newGroup.Gname + "\" with owner id: " + strconv.FormatUint(uuid, 10))

			buf, err := json.Marshal(newGroup)
			if err != nil {
				return err
			}

			w.WriteString(idStr + "\n")
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

	// get details for group to add to
	gid, _ := r.ReadString('\n')
	gid = strings.TrimSpace(gid)
	uintGid, err := strconv.ParseUint(strings.TrimSpace(gid), 10, 64)
	if err != nil {
		fmt.Println("\t\tNot a uint or gid")
		w.WriteByte(2)
		w.Flush()
		return nil
	}

	err = md.userDB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("groups"))

		v := b.Get(itob(uintGid))

		if v == nil {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("\t\tBad access")
		}

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		if tmpGroup.Owner != uuid {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("\t\tBad access")
		}

		// authorised and group exists
		w.WriteByte(1)
		w.Flush()

		return nil
	})

	if err != nil {

		fmt.Println("\tInvalid access to group: " + strings.TrimSpace(gid))
		return nil
	}

	var users []uint64

	// read in all users
	for i := 2; i < int(lenArgs); i++ {

		user, _ := r.ReadString('\n')
		user = strings.TrimSpace(user)

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

		newUsers := ""

		for _, u := range users {

			if !utils.Contains(u, tmpGroup.Members) {
				tmpGroup.Members = append(tmpGroup.Members, u)
				fmt.Printf("\tAdding user: %d to group %d\n", u, tmpGroup.Gid)
				newUsers = newUsers + strconv.FormatUint(u, 10) + ", "
			}
		}

		buf, err := json.Marshal(tmpGroup)
		if err != nil {
			return err
		}

		var anGroup utils.Group
		json.Unmarshal(buf, &anGroup)

		w.WriteString(newUsers + "\n")
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

	// get details for group to remove from
	gid, _ := r.ReadString('\n')
	gid = strings.TrimSpace(gid)
	uintGid, err := strconv.ParseUint(strings.TrimSpace(gid), 10, 64)
	if err != nil {
		fmt.Println("\t\tNot a uint or gid")
		w.WriteByte(2)
		w.Flush()
		return nil
	}

	// ensure valid user and group
	err = md.userDB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("groups"))

		v := b.Get(itob(uintGid))

		if v == nil {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("\t\tBad access")
		}

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		if tmpGroup.Owner != uuid {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("\t\tBad access")
		}

		// authorised and group exists
		w.WriteByte(1)
		w.Flush()

		return nil
	})

	if err != nil {

		fmt.Println("\tInvalid access to group: " + strings.TrimSpace(gid))
		return nil
	}

	var users []uint64

	// read in all users
	for i := 2; i < int(lenArgs); i++ {

		user, _ := r.ReadString('\n')
		user = strings.TrimSpace(user)

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

					fmt.Printf("\tRemoving user: %d from group %d\n", u, tmpGroup.Gid)

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

		w.WriteString(removedUsers + "\n")
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
		fmt.Println("\t\tNot a uint or gid")
		w.WriteByte(2)
		w.Flush()
		return nil
	}

	var members []uint64

	// ensure valid group
	err = md.userDB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("groups"))

		v := b.Get(itob(uintGid))

		if v == nil {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("\t\tBad access")
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

		fmt.Println("\tInvalid access to group: " + gid)
		return nil
	}

	result := ""

	for i, member := range members {

		fmt.Printf("\tMember no. %d = uuid of %d\n", i, member)
		result = result + strconv.FormatUint(member, 10) + ", "
	}

	w.WriteString(result + "\n")
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
		fmt.Println("\t\tNot a uint or gid")
		w.WriteByte(2)
		w.Flush()
		return nil
	}

	// add users to the group in the database
	md.userDB.Update(func(tx *bolt.Tx) (err error) {

		b := tx.Bucket([]byte("groups"))

		v := b.Get(itob(uintGid))

		if v == nil {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("\t\tBad access: not a group")
		}

		var tmpGroup utils.Group
		json.Unmarshal(v, &tmpGroup)

		if tmpGroup.Owner != uuid {

			w.WriteByte(2)
			w.Flush()
			return fmt.Errorf("\t\tBad access: not owner")
		}

		// authorised and group exists
		w.WriteByte(1)
		w.Flush()

		fmt.Printf("\tdeleting group: %d\n", tmpGroup.Gid)

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
			exists = true
		}

		return nil
	})
	return
}

func listGroups(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	mod, _ := r.ReadByte()

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

		}

		w.WriteByte(2)
		w.Flush() // indicate to client that we are done listing

		return nil
	})

	return
}

func checkFile(uuid uint64, targetPath, mod string, md *MDService) (auth bool) {

	//fmt.Println("Getting filestats for: " + path.Join(md.getPath(), "files", targetPath))
	_, _, _, owner, groups, permissions, err := getFile(path.Join(md.getPath(), "files", targetPath))
	if err != nil {
		fmt.Println("\t\tNO FILE AT: " + md.getPath() + "files" + targetPath)
		return false
	}

	//fmt.Printf("filestats: %d, %v, %v\n", owner, groups, permissions)

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

	//fmt.Printf("%b == %b, %b\n", auth, base, permissions[1])

	return base && auth
}

func checkBase(uuid uint64, targetPath, mod string, md *MDService) (auth bool) {

	basePath := strings.TrimSuffix(targetPath, "/"+path.Base(targetPath))
	//fmt.Println("Checking basePath: " + basePath)
	return checkEntry(uuid, basePath, mod, md)
}

func checkEntry(uuid uint64, targetPath, mod string, md *MDService) (auth bool) {

	fmt.Printf("\tChecking %s permissions for user %d on \"%s\"\n", mod, uuid, targetPath)

	// check all the d in dirs for Xecute
	dirs := strings.Split(targetPath, "/")
	if targetPath == "/" || targetPath == "" {
		fmt.Println("\tRoot directory access permitted")
		return true
	}

	traverser := ""
	auth = true

	for i, d := range dirs {

		traverser = path.Join("/", traverser, d)

		if i == len(dirs)-1 {

			owner, groups, permissions, err := getPerm(md.getPath() + "files/" + traverser + "/")
			if err != nil {
				fmt.Println("\t\tError: no .perm file exists at \"" + md.getPath() + "files" + traverser + "/.perm\"")
				fmt.Printf("\tPermission denied for user %d\n", uuid)
				return false
			}

			hasGroup := false
			if owner != uuid {
				auth = false

			}
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

			switch mod {
			case "r":
				auth = auth || (hasGroup && permissions[0]) || permissions[3]

			case "w":
				auth = auth || (hasGroup && permissions[1]) || permissions[4]

			case "x":
				auth = auth || (hasGroup && permissions[2]) || permissions[5]
			}
		} else if i != 0 {

			owner, groups, permissions, err := getPerm(md.getPath() + "files/" + traverser + "/")
			if err != nil {
				fmt.Println("\t\tError: no .perm file exists at \"" + md.getPath() + "files" + traverser + "/.perm\"")
				fmt.Printf("\tPermission denied for user %d\n", uuid)
				return false
			}

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

			if !auth {
				fmt.Printf("\tPermission denied imm for user %d\n", uuid)
				return auth

			}
		}
	}

	if !auth {
		fmt.Printf("\tPermission denied for user %d\n", uuid)
	} else {
		fmt.Printf("\tPermission granted for user %d\n", uuid)
	}
	//fmt.Printf("Auth = %b, Traverser = %s\n", auth, traverser)
	return auth
}

func permit(uuid uint64, conn net.Conn, r *bufio.Reader, w *bufio.Writer, md *MDService) (err error) {

	currentDir, _ := r.ReadString('\n')
	currentDir = strings.TrimSpace(currentDir)

	flag, _ := r.ReadString('\n')
	flag = strings.TrimSpace(flag)
	if flag == "INV" {
		fmt.Printf("\t\tInvalid flag from user %d\n", uuid)
		return nil
	}
	lenArgs, _ := r.ReadByte()

	targetPath, _ := r.ReadString('\n')
	targetPath = strings.TrimSpace(targetPath)

	if !path.IsAbs(targetPath) {
		targetPath = path.Join(currentDir, targetPath)
	}

	fmt.Printf("\tUser %d called permit %s on: \"%s\"\n", uuid, flag, targetPath)

	src, err := os.Stat(md.getPath() + "files" + targetPath)
	if err != nil || utils.IsHidden(targetPath) {
		// exists, not hidden path

		fmt.Printf("\tUser %d call to permit %s on: \"%s\" was invalid: target does not exist\n", uuid, flag, targetPath)
		return nil
	}

	var groups []uint64

	if src.IsDir() {
		addPerms, _ := r.ReadString('\n')
		addPerms = strings.TrimSpace(addPerms)

		for i := 4; i < int(lenArgs); i++ {
			group, _ := r.ReadString('\n')
			gid, err := strconv.ParseUint(strings.TrimSpace(group), 10, 64)
			if err != nil {
				continue
			}
			groups = append(groups, gid)
		}

		if checkEntry(uuid, targetPath, "owner", md) {

			owner, existingGroups, permissions, err := getPerm(md.getPath() + "files" + targetPath)
			if err != nil {
				fmt.Println("\t\tError: no .perm file exists at \"" + md.getPath() + "files" + targetPath + "/.perm\"")
				return nil
			}

			switch flag {
			case "-g":
				fmt.Printf("\tPermitting groups ")
				for _, g := range groups {

					if !utils.Contains(g, existingGroups) {
						fmt.Printf("%d, ", g)
						existingGroups = append(existingGroups, g)
					}
				}
				fmt.Print("for: ")
				if strings.Contains(addPerms, "r") {
					permissions[0] = true
					fmt.Print("r")
				}
				if strings.Contains(addPerms, "w") {
					permissions[1] = true
					fmt.Print("w")

				}
				if strings.Contains(addPerms, "x") {
					permissions[2] = true
					fmt.Print("x")

				}
				fmt.Println(" to " + targetPath)

			case "-w":
				fmt.Printf("\tPermitting world for: ")
				if strings.Contains(addPerms, "r") {
					permissions[3] = true
					fmt.Print("r")

				}
				if strings.Contains(addPerms, "w") {
					permissions[4] = true
					fmt.Print("w")

				}
				if strings.Contains(addPerms, "x") {
					permissions[5] = true
					fmt.Print("x")

				}
				fmt.Println(" to " + targetPath)
			}

			err = createPerm(md.getPath()+"files"+targetPath, owner, existingGroups, permissions)
			if err != nil {
				fmt.Println("\t\tError: there was an issue setting .perm for \"" + md.getPath() + "files" + targetPath + "/.perm\"\n\t\tCould not set new permissions")
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
				fmt.Println("\t\tError: no permissions could be read for file: \"" + md.getPath() + "files" + targetPath + "\"\n\t\tCould not set new permissions")
				return nil
			}

			switch flag {
			case "-g":
				fmt.Printf("\tPermitting groups ")

				for _, g := range groups {

					if !utils.Contains(g, existingGroups) {
						existingGroups = append(existingGroups, g)
						fmt.Printf("%d, ", g)
					}
				}
				fmt.Println("for request access to " + targetPath)

				permissions[0] = true

			case "-w":
				fmt.Println("\tPermitting world for request access to " + targetPath)
				permissions[1] = true
			}

			err = createFile(md.getPath()+"files"+targetPath, hash, stnode, protected, owner, existingGroups, permissions)
			if err != nil {
				fmt.Println("\t\tError re-writing permissions for file: " + targetPath)
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
		fmt.Printf("\t\tInvalid flag from user %d\n", uuid)
		return nil
	}
	lenArgs, _ := r.ReadByte()

	targetPath, _ := r.ReadString('\n')
	targetPath = strings.TrimSpace(targetPath)

	if !path.IsAbs(targetPath) {
		targetPath = path.Join(currentDir, targetPath)
	}

	fmt.Printf("\tUser %d called deny %s on: \"%s\"\n", uuid, flag, targetPath)

	src, err := os.Stat(md.getPath() + "files" + targetPath)
	if err != nil || utils.IsHidden(targetPath) {
		// exists, not hidden path

		fmt.Printf("\tUser %d call to deny %s on: \"%s\" was invalid: target does not exist\n", uuid, flag, targetPath)
		return nil
	}

	var groups []uint64

	if src.IsDir() {
		addPerms, _ := r.ReadString('\n')
		addPerms = strings.TrimSpace(addPerms)

		for i := 4; i < int(lenArgs); i++ {
			group, _ := r.ReadString('\n')
			gid, err := strconv.ParseUint(strings.TrimSpace(group), 10, 64)
			if err != nil {
				continue
			}
			groups = append(groups, gid)
		}

		if checkEntry(uuid, targetPath, "owner", md) {

			owner, existingGroups, permissions, err := getPerm(md.getPath() + "files" + targetPath)
			if err != nil {
				fmt.Println("\t\tError: no .perm file exists at \"" + md.getPath() + "files" + targetPath + "/.perm\"")
				return nil
			}

			switch flag {
			case "-g":
				fmt.Printf("\tDenying groups ")
				for _, g := range groups {
					for i, gr := range existingGroups {
						if g == gr {
							existingGroups = append(existingGroups[:i], existingGroups[i+1:]...)
							fmt.Printf("%d, ", g)
							break
						}
					}
				}
				fmt.Print("for: ")

				if strings.Contains(addPerms, "r") {
					permissions[0] = false
					fmt.Print("r")

				}
				if strings.Contains(addPerms, "w") {
					permissions[1] = false
					fmt.Print("w")

				}
				if strings.Contains(addPerms, "x") {
					permissions[2] = false
					fmt.Print("x")

				}
				fmt.Println(" to " + targetPath)

			case "-w":

				fmt.Printf("\tDenying world for: ")

				if strings.Contains(addPerms, "r") {
					permissions[3] = false
					fmt.Print("r")

				}
				if strings.Contains(addPerms, "w") {
					permissions[4] = false
					fmt.Print("w")

				}
				if strings.Contains(addPerms, "x") {
					permissions[5] = false
					fmt.Print("x")

				}
				fmt.Println(" to " + targetPath)
			}

			err = createPerm(md.getPath()+"files"+targetPath, owner, existingGroups, permissions)
			if err != nil {
				fmt.Println("\t\tError: there was an issue setting .perm for \"" + md.getPath() + "files" + targetPath + "/.perm\"\n\t\tCould not set new permissions")
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
				fmt.Println("\t\tError: no permissions could be read for file: \"" + md.getPath() + "files" + targetPath + "\"\n\t\tCould not set new permissions")
				return nil
			}

			switch flag {
			case "-g":
				fmt.Printf("\tDenying groups ")

				for _, g := range groups {
					for i, gr := range existingGroups {
						if g == gr {
							existingGroups = append(existingGroups[:i], existingGroups[i+1:]...)
							fmt.Printf("%d, ", g)
							break
						}
					}
				}
				fmt.Println("for request access to " + targetPath)

				permissions[0] = false

			case "-w":
				fmt.Println("\tDenying world for request access to " + targetPath)
				permissions[1] = false
			}

			err = createFile(md.getPath()+"files"+targetPath, hash, stnode, protected, owner, existingGroups, permissions)
			if err != nil {
				fmt.Println("\t\tError re-writing permissions for file: " + targetPath)
			}
		}
	}
	return nil
}
