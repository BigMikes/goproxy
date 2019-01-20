package main

import (
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", "localhost:4040")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil { //TODO: check type and retry in case of temporary err
			log.Fatal(err)
		}
		go proxy(conn)
	}
}
