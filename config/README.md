#Config Package
This package contains structs and functions to handle the configuration of the servers. It also has sample configuration files in JSON format.

##config.go
This file contains all the structs and functions that are need for handling configuration.

####Configuration struct
This struct contains strings to represent the following:
- the protocol being used for communication
- the host
- the port it should listen on
- the path it should store files to
- a unique ID

```Go
type Configuration struct {
	Protocol string
	Host     string
	Port     string
	Path     string
	Unid     string
}
```

####ParseConfiguration() function
This function takes a path to a file and returns a Configuration object. The file should be a JSON file as it gets parsed into a Configuration object which is then returned.

```Go
// get configuration information from JSON file
func ParseConfiguration(filename string) Configuration {

	file, _ := os.Open(filename)
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}
```

####SetConfiguration() function
This function takes a Configuration struct and a filepath and decodes the struct into JSON which is saved at the filepath. An error is returned with nil indicating success.

```Go
// saves the passed in configuration to file
func SetConfiguration(conf Configuration, filename string) (err error) {

	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	encoder := json.NewEncoder(fo)
	err = encoder.Encode(conf)
	if err != nil {
		return err
	}
	return err
}
```

##config_test.go
This file contains unit tests for the ParseConfiguration() function in `config.go`. It uses the function to parse a sample configuration file and then compares the member variables of the returned struct against the expected values.

##handling_codes.json
This file is used to keep track of what all the `uint8` byte codes mean to the different pieces of software. For example, to the client the code `1` might mean `send` whereas to the storage node this would mean `receive`.

##server_conf.json
This file contains a sample server configuration.

```JSON
{
	"protocol": "tcp",
	"host": "localhost",
	"port": "8081",
	"path": "/path/to/files/"
}
```