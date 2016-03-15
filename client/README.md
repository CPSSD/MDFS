##client.go
This program is used to interact with a running instance of a metadata service. The program will panic if it tries to connect to a socket where there is no metadata service listening.

###Configuration
The location of the metadata service and the clients username are defined at the beginning of the main function. This will later be moved outside of the file.

```Go
socket := "localhost:1994"
user := "jim"
```

###Usage
The client always begins in the root directory. The current user and directory are displayed to the left of the prompt, separated by a colon.

```
jim:/ >>
```
Here you can see that the user 'jim' is in their root directory (denoted by the `/` symbol).

####Available Commands
The user types in their command followed by the return key just like with most shell programs.

The following commands are currently implemented.

```
ls
cd
pwd
mkdir
rmdir
exit
```

#####Example Usages
```BASH
jim:/ >> ls
bar foo
jim:/ >> ls foo
foo:
bar

jim:/ >> ls foo bar
foo:
a.txt b.txt c.txt

bar:
x.jpg y.wav z.mp4
```

```BASH
jim:/ >> cd foo
jim:/foo >> cd ..
jim:/ >> cd bar
jim:/bar >> cd ..
jim:/ >> cd ..
jim:/ >> cd foobar
Not a directory
jim:/ >>
```

```BASH
jim:/ >> ls

jim:/ >> mkdir foo
jim:/ >> ls
foo
jim:/ >> mkdir bar
jim:/ >> ls
bar foo
```

```BASH
jim:/ >> ls
bar foo
jim:/ >> rmdir foo
jim:/ >> ls
bar
```