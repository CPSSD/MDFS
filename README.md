# MDFS [![Build Status](https://travis-ci.com/CPSSD/MDFS.svg?token=ZNLEp9wQPE3kma4CBH8m&branch=master)](https://travis-ci.com/CPSSD/MDFS)
Massively Distributed File System

## Repository Layout
There are three pieces of software contained within the MDFS repository: the metadata service, the storage node and the client for interacting with the two services. There are currently separate clients for interacting with the two services but these will eventually be merged.

## Setup
### Testing
Before testing, run the entire setup using ``go run testing_files/testing_init.go && go run storagenode/config/stnode_init.go && go run mdservice/config/mdservice_init.go``. This will create the necessary folder structure ``$HOME/.testing_files/``. The test files will be copied to this location. This will also set up the neccessary files as mentioned below.

### Storage Node
Before you begin, setup the storage node using ``go run storagenode/config/stnode_init.go``. This will create the necessary folder structure ``$HOME/.stnode/``. The server's configuration files will be copied to ``$HOME/.stndode/stnode_conf.json``.

### Metadata Service
Setup the metadata service using ``go run mdservice/config/mdservice_init.go``. This establishes a folder to store files for the mdservice itself at ``$HOME/.mdservice/files``. The mdservice's config files are stored under ``$HOME/.mdservice/mdservice_conf.json``.

## Usage
### Run the storage node
The storage node server can then be run
```bash
go run storagenode/stnode.go
```

### Run the metadata service
The metadata service can then be run as follows
```bash
go run mdservice/mdservice.go
```

### Run the client software
The client software is run as follows for the mdservice
```bash
go run client/mdservice/mdservice_client.go
```

### Request a file
Request a file using the flag `-request={hex representation of hash}`

```
go run client/stnode/stnode_client.go -request=6f5902ac237024bdd0c176cb93063dc4
```

File can then be found in /path/to/files/output

### Send a file
Send a file using the flag `-send={path to file}`

```
go run client/stnode/stnode_client.go -send=/path/to/files/test.jpg
```

File will be stored in the configured storage location in the json file (/path/to/files/ by default).

## Using the metadata service client
``go run client/mdservice/mdservice_client.go`` to run interactive shell for interacting with the metadata service. It currently only runs locally to test filesystem creation.

It currently accepts the following bash-like commands:
```BASH
cd
pwd
ls
mkdir
rmdir
exit
```