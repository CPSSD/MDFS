package mdservice

import (
	"net"
	"fmt"
)

// structs
type Node struct {
	parent      *DirNode
	name        string
	permissions uint16
	owner       UUID
}

type FileNode struct {
	Node    // anonymous field of type Node
	hash    []byte
	storage []UNID
}

type DirNode struct {
	Node     // anonymous field of type Node
	contents []iNode
}

type UUID struct {
	username string
}

type UNID struct {
	location net.IP
}


// interfaces
type iNode interface {
	String() string
}

// methods
func (n *Node) initialise(p *DirNode, nm string, perm uint16, ownr *UUID) {
	n.parent = p
	n.name = nm
	n.permissions = perm
	n.owner = *ownr
}

func (n *Node) GetName() string {
	return n.name
}

func (dir *DirNode) Ls() {
	for _, elem := range dir.contents {
		fmt.Println(elem)
	}
}

func MkRoot() *DirNode {
	root := new(UUID)
	root.Initialise("root")
	rootDir := new(DirNode)
	rootDir.initialise(nil, "/", 0, root)
	rootDir.contents = nil
	return rootDir
}

func (dir *DirNode) MkDir(nm string, perm uint16, ownr *UUID) *DirNode {
	d := new(DirNode)
	d.initialise(dir, nm, perm, ownr)
	d.contents = nil
	dir.contents = append(dir.contents, *d)
	return d
}

func (dir *DirNode) IsEmpty() bool {
	if dir.contents == nil {
		return true
	}
	return false
}

func (u *UUID) Initialise(uname string) {
	u.username = uname
}


// string functions
func (u UUID) String() string {
	return fmt.Sprintf("%s", u.username)
}

func (n Node) String() string {
	return fmt.Sprintf("%d\t%v\t%s", n.permissions, n.owner, n.name)
}