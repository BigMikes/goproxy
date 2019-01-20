package main

import (
	"bytes"
	"io"
	"log"
	"net"
)

type proxyedConn struct {
	net.Conn
	io.Reader
}

func (p proxyedConn) Read(b []byte) (int, error) {
	return p.Reader.Read(b)
}

func proxy(conn net.Conn) {
	defer conn.Close()

	//First, read the remote endpoint
	remoteEp, firstReq, err := deriveDestEndpoint(conn)
	if err != nil {
		log.Println(err)
		return
	}

	//Write the first req back into the wire
	var buf bytes.Buffer
	err = firstReq.Write(&buf)
	if err != nil {
		log.Println(err)
		return
	}

	//Create the net.Conn wrapper
	pc := proxyedConn{
		Conn:   conn,
		Reader: io.MultiReader(&buf, conn), //It will first read from buf, where the first req is
	}

	remote, err := net.Dial("tcp", remoteEp)
	if err != nil {
		log.Println(err)
		return
	}
	defer remote.Close()

	log.Println("Proxying conn from", conn.RemoteAddr(), "to", remoteEp)

	go redirectConn(remote, pc)
	redirectConn(pc, remote)
	log.Println("Proxy exited for", conn.RemoteAddr(), "to", remoteEp)
}

func redirectConn(dest, src net.Conn) {
	_, err := io.Copy(dest, src)
	log.Println("From", src, "to", dest, "exited with err = ", err)
}
