package mdservice

import (
	"net"
)

type Node struct {
	parent      *Node
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
	contents []*Node
}

type UUID struct {
	username string
}

type UNID struct {
	location net.IP
}



func (n *Node) initialise(p *Node, nm string, perm uint16, ownr UUID) {
	n.parent := p
	n.name := name
	n.permissions := perm
	n.owner := ownr
}

func InitialiseDir(p *Node, nm string, perm uint16, ownr UUID, conts []*Node) *DirNode {
	dir := new(DirNode)
	dir.initialise(p, nm, perm, ownr)
	dir.contents = conts
}

func (u *UNID) Initialise(uname string) {
	u.username = uname
}
/*
func mkdir() {
	
}

func rmdir() {

}*/