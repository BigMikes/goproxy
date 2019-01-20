package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	firstReadDeadline time.Duration = time.Second * 10
	defaultPort       string        = "80"
)

func deriveDestEndpoint(conn net.Conn) (string, *http.Request, error) {
	conn.SetReadDeadline(time.Now().Add(firstReadDeadline))
	buf := bufio.NewReader(conn)
	req, err := http.ReadRequest(buf)
	if err != nil {
		return "", nil, err
	}
	log.Printf("Received req: %+v\n", req)
	//Reset the deadline
	conn.SetReadDeadline(time.Time{})
	hostname, err := fixHost(req.URL)
	if err != nil {
		return "", nil, err
	}
	return hostname, req, nil
}

// Fixes the host address appending ":80"
// if no protocol is specified already
func fixHost(u *url.URL) (string, error) {
	hostname := u.Hostname()
	port := u.Port()
	if port == "" {
		port = defaultPort //If port unspecified, set it to default value
	}
	return strings.Join([]string{hostname, port}, ":"), nil
}
