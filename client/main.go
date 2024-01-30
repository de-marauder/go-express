package main

import (
	"fmt"
	"io"
	"log"
	"net"
)


const (
	net_protocol = "tcp"
	host         = "localhost"
	port         = "7000"
	buf_size     = 1024
)

// Define a structure for our client
type Client struct {
	protocol string
	addr     string
	conn net.Conn
}

func NewClient(protocol string, address string) *Client {
	conn, err := net.Dial(protocol, address)
	logError(err)
	return &Client{
		protocol: protocol,
		addr: address,
		conn: conn,
	}
}

func main() {
	client := NewClient(net_protocol, host+":"+port)
	handleConnection(client.conn)
}

// Close connection before leaving function scope
// Write message to server via connection's buffer writer
// Read response from server via connection's read buffer
func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Client server running on %s\n", conn.LocalAddr())

	io.WriteString(conn, "Opening connection to server")

	buffer := make([]byte, buf_size)
	_, err := conn.Read(buffer)
	logError(err)

	log.Println(string(buffer))
}

// Log errors and fail
func logError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}