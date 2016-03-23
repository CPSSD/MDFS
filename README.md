# MDFS [![Build Status](https://travis-ci.com/CPSSD/MDFS.svg?token=ZNLEp9wQPE3kma4CBH8m&branch=master)](https://travis-ci.com/CPSSD/MDFS)
Massively Distributed File System

## Repository Layout
There are three pieces of software contained within the MDFS repository: the metadata service, the storage node and the client for interacting with the two services.

## Setup
### Testing
Before testing, run the entire setup using ``go run testing_files/testing_init.go && go run storagenode/config/stnode_init.go && go run mdservice/config/mdservice_init.go``. This will create the necessary folder structure ``$HOME/.testing_files/``. The test files will be copied to this location. This will also set up the neccessary files as mentioned below.

### Storage Node
Before you begin, setup the storage node using ``go run configure.go``. Follow the on screen steps after selecting ``1`` to set up the storage node. This will create the necessary folder structure ``$HOME/.stnode/``. The server's configuration files will be copied to ``$HOME/.mdfs/stndode/.stnode_conf.json``.

### Metadata Service
Setup the metadata service using ``go run configure.go``. Follow the on screen steps after selecting ``2`` to set up the metadata service. This establishes a folder to store files for the mdservice itself at ``$HOME/.mdservice/files``. The mdservice's config files are stored under ``$HOME/.mdfs/mdservice/.mdservice_conf.json``.

## Usage
### Install and run the storage node
The storage node server can then be run
```bash
go install storagenode/stnode.go
stnode
```

### Run the metadata service
The metadata service can then be run as follows
```bash
go install mdservice/mdservice.go
mdservice
```

### Run the client software
The client software is run as follows for the mdservice
```bash
go install client/mdfs_client.go
```

Once the client is run, you will be prompted for a username and the connection details of your metadata service. Once complete, you will be presented with the command prompt for the mdfs-client.

```bash
thewall:2:/ >> 
```

### Available commands
The following BASH-like commands are currently available to users.

```bash
Massively Distributed Filesystem client, version 1.0-demo
The following is a list of available commands and their usage.
All commands will require the appropriate permissions (read, write or execute) to be successful

Usage: ls               [-V] [directory...]
 
       pwd
       cd               [directory]
       mkdir            [directory...] 
       rmdir            [directory...]
      
       send             [filename] [-p [UUID...]]
       request          [filename]
       rm               [filename...]
      
       permit           [-w (directory [rwx] | filename) | -g (directory [rwx] [GID...])]
       deny             [-w (directory [rwx] | filename) | -g (directory [rwx] [GID...])]
      
       create-group     [groupname...]
       delete-group     [GID]
      
       group-add        [GID] [UUID...]
       group-remove     [GID] [UUID...]
       group-ls         [GID]
      
       list-groups      (-m | -o | -mV | -oV | -V)

       exit
```