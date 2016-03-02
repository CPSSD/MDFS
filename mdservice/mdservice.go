package main

import (
	"github.com/CPSSD/MDFS/server"
)

func main() {
	var s server.MDService
	server.Start(&s)
}
