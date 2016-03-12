package main

import (
	"fmt"
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/mdservice"
	"os"
)

func main() {

	fmt.Println("This is the installation program for the MDFS")
	fmt.Println("Please select which software you would like to install")
	fmt.Println("1) Client software")
	fmt.Println("2) Storage node software")
	fmt.Println("3) Metadata service")

	// get user selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter selection: ")
	selection, _ := reader.ReadString('\n')

	switch selection {
	case "1": // configure client software
		continue

	case "2": // configure storage node
		path, port := getConfig("files")
		err := setup(path, port, "./storagenode/config/", "stnode_conf.json")
		if err != nil {
			panic(err)
		}

		fmt.Println("Storage node has been initialised.")
		fmt.Printf("The configuration file can be found at %s.stnode_conf.json.\n", path)

	case "3": // configure metadata service
		path, port := getConfig("metadata")
		err := setup(path, port, "./mdservice/config/", "mdservice_conf.json")
		if err != nil {
			panic(err)
		}

		// create a subdirectory called files
		err = os.MkdirAll(path+"files/", 0700)
		if err != nil {
			panic(err)
		}

		fmt.Println("Metadata service has been initialised.")
		fmt.Printf("The configuration file can be found at %s.mdservice_conf.json.\n", path)
	}
}

func getConfig(string store) (path string, port string) {

	// get desired storage location
	fmt.Printf("Please enter the absolute path to the directory in which you would like to store the %s.\n", store)
	fmt.Println("Please ensure you include the trailing directory slash")
	fmt.Print("Path: ")
	path, _ = reader.ReadString('\n')

	fmt.Println("Please enter the port you want the service to listen on.")
	fmt.Println("Port: ")
	port, _ = reader.ReadString('\n')

	return path, port
}

func setup(path string, port string, sample string, fname string) error {

	// create the supplied file path
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return err
	}

	// get the sample configuration file from the repo
	conf := config.ParseConfiguration(sample + fname)
	conf.Path = path // change the path variable
	conf.Port = port // change the port variable

	// encode the new object to a json file
	err = config.SetConfiguration(conf, path+fname)
	return err
}
