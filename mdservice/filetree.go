package filetree

import (
	"net"
)

type Node struct {
	parent      Node
	name        string
	permissions uint16
	owner       UUID
}

type FileNode struct {
	Node
	hash    []byte
	storage []UNID
}

type DirNode struct {
	Node
	contents []Node
}

type UUID struct {
	username string
}

type UNID struct {
	location net.IP
}
