package mdservice

import (
	"net"
	"errors"
	"fmt"
	"strings"
	"reflect"
	"github.com/CPSSD/MDFS/utils"
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
	isFileNode() bool
	isDirNode() bool
	GetName() string
}



// node methods
func (n *Node) initialise(p *DirNode, nm string, perm uint16, ownr *UUID) {
	n.parent = p
	n.name = nm
	n.permissions = perm
	n.owner = *ownr
}

func (n Node) GetName() string {
	return n.name
}

func (n Node) isFileNode() bool {
	return reflect.TypeOf(n).String() == "FileNode"
}

func (n Node) isDirNode() bool {
	return reflect.TypeOf(n).String() == "DirNode"
}



// directory methods
func (dir *DirNode) Ls() {
	for _, elem := range dir.contents {
		fmt.Println(elem)
	}
}

func (dir *DirNode) Pwd() {
	// dir node pointer for traversing the path back to root
	traverser := *dir

	// buffer to hold path
	buffer := []string{}

	for i := traverser.GetName(); i != "/"; i = traverser.GetName() {
		buffer = utils.Prepend(buffer, i)
		traverser = *traverser.parent
	}
	fmt.Println("/"+strings.Join(buffer, "/"))
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

func MkRoot() *DirNode {
	root := new(UUID)
	root.Initialise("root")
	rootDir := new(DirNode)
	rootDir.initialise(nil, "/", 0, root)
	rootDir.contents = nil
	return rootDir
}

func Cd(current *DirNode, next string) (error, *DirNode) {
	if next == ".." && current.name != "/" {
		return nil, current.parent
	}

	// check if requested dir is child of current
	for _, child := range current.contents {
		v, ok := child.(DirNode) // assert that the child is a dir and not a file
		if ok && child.GetName() == next {
			return nil, &v
		}
	}
	err := errors.New("Directory does not exist")
	return err, current
}



// uuid methods
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