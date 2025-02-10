package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

type client struct {
    conn net.Conn
    wC chan []byte
    sWC chan []byte
}
type server struct {
    conn net.Conn
    disconnected chan struct{}
    sWC <- chan []byte
    mutex *sync.RWMutex
    clients *map[client]struct{}
}
var clients map[client]struct{}
var mutex sync.RWMutex
func main() {
    if len(os.Args) != 3 {
        usage()
        os.Exit(1)
    }
    l, err := net.Listen("tcp",os.Args[1])
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    sCon, err := net.Dial("tcp", os.Args[2])
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    clients = make(map[client]struct{})
    sWC := make(chan []byte)
    srv := server{
        conn: sCon,
        sWC: sWC,
        disconnected: make(chan struct{}),
        clients: &clients,
        mutex: &mutex,
    }
    go srv.handleServer()
    for {
        con, err := l.Accept()
        if err != nil {
            fmt.Fprintln(os.Stderr,err)
            continue
        }
        c := client{
            conn: con,
            wC: make(chan []byte),
            sWC: sWC,
        }
        mutex.Lock()
        fmt.Println("Client connected")
        clients[c] = struct{}{}
        mutex.Unlock()
        go handleClient(c)
    }
}

func handleClient(cCon client) {
    go func(r io.Reader) {
        buff := make([]byte, 1024)
        for {
            n, err := r.Read(buff)
            if err == io.EOF {
                return
            }
            cCon.sWC <- buff[:n]
        }
    }(cCon.conn)
}

func (srv *server) handleServer() {
    go srv.read()
    for {
        select {
        case d := <- srv.sWC:
            srv.conn.Write(d)
        case <- srv.disconnected:
            srv.reconnect()
            go srv.read()
        }
    }
}
func (srv *server) reconnect()  {
    fmt.Println("Connection to server lost")
    for {
        fmt.Println("Trying to reconnect")
        c, err := net.Dial("tcp",srv.conn.RemoteAddr().String())
        if err == nil {
            srv.conn = c
            fmt.Println("reconnected")
            return
        }
        <- time.After(2*time.Second)
    }
}
func (srv *server) read()  {
    buff := make([]byte, 1024)
    for {
        n, err := srv.conn.Read(buff)
        if err == io.EOF {
            srv.disconnected <- struct{}{} 
            return
        }
        srv.mutex.Lock()
        for c := range *srv.clients {
            nc, _ := c.conn.Write(buff[:n])
            if nc == 0 {
                fmt.Println("removing client, disconnected")
                delete(clients, c)
            }
        }
        srv.mutex.Unlock()
    }
}


func usage()  {
    fmt.Printf("Usage: %s <listen:port> <destination:port>",os.Args[0])
    fmt.Printf("Example: %s 127.0.0.1:4445 127.0.0.1:4444 connects to 127.0.0.1 at port 4444 and listens on 4445",os.Args[0])
}


