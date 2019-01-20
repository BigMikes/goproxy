package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
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

func proxy(conn net.Conn) {
	//TODO add timeouts where appropriate
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

	pc := proxyedConn{
		Conn:   conn,
		Reader: io.MultiReader(&buf, conn),
	}

	remote, err := net.Dial("tcp", remoteEp+":80")
	if err != nil {
		log.Println(err)
		return
	}
	defer remote.Close()

	log.Println("Proxying conn from", conn.RemoteAddr(), "to", remoteEp)

	go io.Copy(remote, pc)
	io.Copy(pc, remote)
	log.Println("Proxy exited for", conn.RemoteAddr(), "to", remoteEp)
}

type proxyedConn struct {
	net.Conn
	io.Reader
}

func (p proxyedConn) Read(b []byte) (int, error) {
	return p.Reader.Read(b)
}

func deriveDestEndpoint(conn net.Conn) (string, *http.Request, error) {
	buf := bufio.NewReader(conn)
	req, err := http.ReadRequest(buf)
	if err != nil {
		return "", nil, err
	}
	return req.Host, req, nil
}

func copyToStderr(conn net.Conn) {
	defer conn.Close()
	for {
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))
		var buf [128]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			log.Printf("Finished with error %v", err)
			return
		}
		os.Stderr.Write(buf[:n])
	}

}
