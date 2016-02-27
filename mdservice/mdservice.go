package mdservice

import (
    "github.com/CPSSD/MDFS/server"
)

func main() {
    var s server.MDService
    server.Start(&s)
}