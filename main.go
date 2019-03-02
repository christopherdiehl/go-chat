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
		server := CreateServer(":8080")
		err := server.Listen()
		fmt.Println(err)
		fmt.Println("Should be listening on port 8080")
	default:
		fmt.Println("Please specify connect or start")
		os.Exit(1)
	}
}

// Client handles the client interaction
type Client struct {
	connection net.Conn    // actual network connection
	outbound   chan string // messages to send to client
	incoming   chan string // messages from client
	username   string      // username
	joined     time.Time   // time joined
}

// Server contains all the server logic
type Server struct {
	port        string
	started     time.Time
	outbound    chan string
	connections []*Client
}

// CreateServer with default values
func CreateServer(port string) *Server {
	return &Server{
		port:        port,
		outbound:    make(chan string, 1),
		connections: make([]*Client, 1),
		started:     time.Now(), // default to now
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

// AddClient appends client to server connections if they possess a unique username
func (server *Server) AddClient(client *Client) error {
	for _, c := range server.connections {
		if c.username == client.username {
			client.outbound <- "Username in use, please select new one"
		}
	}
	server.connections = append(server.connections, client)
	return nil
}

// Broadcast sends message to all the clients
func (server *Server) Broadcast(message string, sender string) error {
	for _, c := range server.connections {
		if c.username != sender {
			c.outbound <- fmt.Sprintf("[%s][%s]: %s", sender, time.Now().String(), message)
		}
	}
	server.outbound <- message
	return nil
}

// HandleClient connection.
// If incoming message - broadcasts
func (server *Server) HandleClient(conn net.Conn) error {
	defer conn.Close()
	incoming := readConnection(conn)
	var username string = ""
	for {
		select {
		case message, open := <-incoming:
			if open == false && username != "" {
				server.Broadcast("left", username)
				return nil
			}
			if username == "" {
				username = message
				server.Broadcast("joined", username)
				continue
			}
			server.Broadcast(message, username)
		case message := <-server.outbound:
			conn.Write([]byte(message))
		}
	}
}

// returns string channel for incoming messages
func readConnection(conn net.Conn) (incoming chan string) {
	incoming = make(chan string, 1)
	go func() {
		for {
			buf := make([]byte, 1028)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println(err)
				close(incoming)
			}
			if n > 1 {
				incoming <- string(buf[:n])
			}
			// time.Sleep(100 * time.Millisecond) // sleep for 100 milliseconds
		}
	}()
	return
}
