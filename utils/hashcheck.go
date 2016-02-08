package utils

import "os"

func CheckForHash(path, hash string) bool {
	if _, err := os.Stat(path + hash); err == nil {
		return true
	} else {
		return false
	}
}
