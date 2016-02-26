package main

import (
	"github.com/CPSSD/MDFS/server"
)

func main() {
	var s server.StorageNode
	server.Start(&s)
}