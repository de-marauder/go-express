package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	net_protocol = "tcp"
	host         = "localhost"
	port         = "7000"
	buf_size     = 1024
)

// Define a structure for our server
type Server struct {
	protocol string
	addr     string
	listener net.Listener
}

// Create a constructor function for the Server
func NewTCPServer(protocol string, addr string) *Server {
	ln, err := net.Listen(protocol, addr)
	logError(err)
	return &Server{
		protocol: protocol,
		addr:     addr,
		listener: ln,
	}
}

// Start a new tcp server that listens on the specified address
// Tell the listener to close just before the function scope is left
// Start an accept loop to receive requests
func main() {
	server := NewTCPServer(net_protocol, host+":"+port)
	defer server.listener.Close()
	fmt.Println("TCP Server listening on", server.addr, "...")
	acceptLoop(server.listener)
	fmt.Printf("Hey!\n")
}

// Create a while loop to accept all incoming connections on a listener
// Handle all accepted connections as goroutines
func acceptLoop(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		logError(err)
		go handleConnection(conn)
	}
}

// Close the connection before leaving function scope
// Make a buffer to accept incoming message. The buffer is currently 1024 but it can be a large as required
// Display message in the console
// Relay response to HTTP response if client is an HTTP client
// Else send regular string
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Accept request
	buf := make([]byte, buf_size)
	n, err := conn.Read(buf)
	logError(err)
	message := string(buf[:n])
	log.Println("Connection from ", conn.RemoteAddr())
	fmt.Printf("%s\n", message)

	// Respond
	if isHTTPRequest(message) {
		// handleHTTPRequest(conn, message)
		const response = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello, Client!\r\n"
		_, err := conn.Write([]byte(response))
		logError(err)
	} else {
		response := "Thanks for connecting"
		_, err = conn.Write([]byte(response))
		logError(err)
	}
}

// Check if message is an HTTP request
func isHTTPRequest(message string) bool {
	scheme := strings.Split(message, "\r")[0]
	return strings.Contains(scheme, "HTTP")
}

func logError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// Identify HTTP verb
// Identify route
// select route handler
// process request
// func handleHTTPRequest(conn net.Conn, message string) {
// 	msgSlice := strings.Split(message, "\r")
// 	scheme := strings.Split((msgSlice[0]), " ")
// 	headers := msgSlice[1 : len(msgSlice)-1]
// 	verb := strings.Trim(scheme[0], " ")
// 	route := strings.Trim(scheme[1], " ")

// 	fmt.Println(verb)
// 	fmt.Println(route)
// 	fmt.Println(headers)

// 	const response = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello, Client!\r\n"
// 	_, err := conn.Write([]byte(response))
// 	logError(err)
// }
