package main

import (
	"github.com/CPSSD/MDFS/mdservice"
	"os"
	"fmt"
	"bufio"
	"strings"
)

var currentDir mdservice.DirNode
var user mdservice.UUID

func main() {
	rootDir := mdservice.MkRoot()
	currentDir := rootDir
	user.Initialise("jim")

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		cmd, _ := reader.ReadString('\n')
		// remove trailing newline character before splitting
		args := strings.Split(strings.TrimSpace(cmd), " ")

		switch (args[0]) {
		case "":
			continue

		case "pwd":
			currentDir.Pwd()

		case "cd":
			err, next := mdservice.Cd(currentDir, args[1])
			if err != nil {
				fmt.Println(err)
			}
			currentDir = next

		case "ls":
			if !currentDir.IsEmpty() {
				fmt.Printf("%s\t%s\t%s\n", "perm", "owner", "name")
				currentDir.Ls()
			}
		
		case "mkdir":
			if len(args) > 1 && args[1] != "" {
				currentDir.MkDir(args[1], 0, &user)
			}
		
		case "exit":
			os.Exit(1)
		
		default:
			fmt.Println("Unrecognised command")
		}
	}
}