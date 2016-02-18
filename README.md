# MDFS https://travis-ci.com/CPSSD/MDFS.svg?token=ZNLEp9wQPE3kma4CBH8m&branch=master
Massively Distributed File System

## Usage
Run the server using ``go run server.go``

### Request a file
Request a file using the flag `-request={hex representation of hash}`

```
go run client -request=6f5902ac237024bdd0c176cb93063dc4
```

File can then be found in /path/to/files/client/output

### Send a file
Send a file using the flag `-send=/path/to/file`

```
go run client -send=/path/to/files/test.jpg
```

File will be stored in the configured storage location in the json file 
