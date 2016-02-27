package utils

import (
	"os/user"
)

func GetUserHome() (string) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return dir
}