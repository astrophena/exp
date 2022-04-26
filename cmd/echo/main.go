// Command echo is a simple TCP server that echoes anything is receives.
package main

import (
	"flag"
	"io"
	"log"
	"net"
)

func main() {
	log.SetFlags(0)

	addr := flag.String("addr", "localhost:3000", "Listen on `host:port`.")
	flag.Parse()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Accepting connection failed: %v", err)
			conn.Close()
			continue
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	log.Printf("Accepted connection from %v.", conn.RemoteAddr())
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				// Probably disconnected.
				log.Printf("%v disconnected.", conn.RemoteAddr())
				return
			}
			log.Printf("conn.Read: %v", err)
			return
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			log.Printf("Write to %v failed: %v", conn.RemoteAddr(), err)
			return
		}
	}
}
