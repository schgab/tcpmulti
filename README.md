# What is this?
tcpmulti is a small go application that connects to a host and listens on a given address and port accepting tcp connections.
Data received by the host is then forwarded to each connection and data received by a connection is forwarded to the host. 

This is useful when you have some application which accepts only a single tcp connection, but you need it to handle multiple connections.

# Build & Run
```
git clone https://github.com/schgab/tcpmulti.git
cd tcpmulti
go build -o tcpmulti main.go
./tcpmulti <local_addr>:<port> <remote_addr>:<port> 
```
