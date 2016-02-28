package client

import (
	"github.com/CPSSD/MDFS/mdservice"
)

func main() {
	rootDir := initialise()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')

		switch (text) {
		case "touch":
		}
	}
}

func initialise() *mdservice.DirNode {
	root := new(mdservice.UUID)
	root.Init("root")
	rootDir := new(mdservice.DirNode)
	rootDir.Init(nil, "root", "0", root)
}