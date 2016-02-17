package main

import (
	"MDFS/utils"
)

func main() {
	
	//test two files for encryption and then decryption
	source := "/path/to/files/input.txt"
	encryp := "/path/to/files/input.enc"
	result := "/path/to/files/result.txt"

 	utils.EncryptFile(source, encryp)
	utils.DecryptFile(encryp, result)

	source = "/path/to/files/image.jpg"
	encryp = "/path/to/files/image.enc"
	result = "/path/to/files/result.jpg"

	utils.EncryptFile(source, encryp)
	utils.DecryptFile(encryp, result)
}