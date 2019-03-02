package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n connect -- connect to chat server\n start -- spool up chat server")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "connect":
		fmt.Println("Connecting you now")
	case "start":
		fmt.Println("Spooling up now")
	default:
		fmt.Println("Please specify connect or start")
		os.Exit(1)
	}
}

// Client handles the client interaction
type Client struct {
	connection net.Conn
	outgoing   *chan string
	incoming   *chan string
	username   string
}

// Server contains all the server logic
type Server struct {
	port    string
	started time.Time
	clients []*Client
}

// CreateServer with default values
func CreateServer(port string) *Server {
	return &Server{
		port:    port,
		started: time.Now(), // default to now
		clients: make([]*Client, 0),
	}
}

// Listen on specified server port
func (server *Server) Listen() error {
	ln, err := net.Listen("tcp", server.port)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection at ", time.Now())
		}
		go server.HandleClient(conn)
	}
}
func (server *Server) Broadcast(message string) error {
	return nil
}

// HandleClient connection.
// If incoming message - broadcasts
func (server *Server) HandleClient(conn net.Conn) error {
	for {
		buf := make([]byte, 1028)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from client")
		}
		message := string(buf[:n])
		_ = message
	}
	return nil
}
