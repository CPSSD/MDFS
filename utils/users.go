package utils

import (
	"os/user"
)

func GetUserHome() (string, err) {
	usr, err := user.Current()
	dir := usr.HomeDir
	return (dir, err)
}