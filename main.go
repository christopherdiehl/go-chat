package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n port to run on")
		os.Exit(1)
	}
	server := CreateServer("127.0.0.1:" + os.Args[1])
	err := server.Listen()
	fmt.Println(err)
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
		connections: make([]*Client, 0),
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
			continue
		}
		go server.HandleClient(conn)
	}
}

// AddClient appends client to server connections if they possess a unique username
func (server *Server) AddClient(client *Client) error {
	for _, c := range server.connections {
		if c.username == client.username {
			client.outbound <- "Username in use, please select new one"
			return errors.New("Username in use")
		}
	}
	server.connections = append(server.connections, client)
	server.Broadcast("joined \n", client.username)
	return nil
}

// Broadcast sends message to all the clients
func (server *Server) Broadcast(message string, sender string) error {
	for _, c := range server.connections {
		if c.username != sender {
			c.outbound <- fmt.Sprintf("[%s][%s]: %s", sender, time.Now().Format(time.UnixDate), message)
		}
	}
	return nil
}

// RemoveClient removes client from server connections
// sends a message to the rest of the chat board letting them know the client has left
// returns true if successful, false if not found
func (server *Server) RemoveClient(client *Client) bool {
	for i, c := range server.connections {
		if c.username == client.username {
			server.Broadcast("left at "+time.Now().Format(time.UnixDate)+"\n", client.username)
			// zero out pointer as per https://github.com/golang/go/wiki/SliceTricks
			copy(server.connections[i:], server.connections[i+1:])
			server.connections[len(server.connections)-1] = nil
			server.connections = server.connections[:len(server.connections)-1]
			return true
		}
	}
	return false
}

// HandleClient connection.
// If incoming message - broadcasts
func (server *Server) HandleClient(conn net.Conn) error {
	defer conn.Close()
	incoming := readConnection(conn)
	conn.Write([]byte("Please enter a username "))
	var username string
	client := &Client{
		outbound:   make(chan string, 1),
		username:   "",
		joined:     time.Now(),
		connection: conn,
		incoming:   incoming,
	}
	for {
		select {
		case message, open := <-incoming:
			// client connection closed
			if open == false {
				// user already in the system, need to remove
				if username != "" {
					server.RemoveClient(client)
				}
				return nil
			}
			if username == "" {
				client.username = strings.Trim(message, "\n")
				if err := server.AddClient(client); err == nil {
					conn.Write([]byte("Welcome, " + client.username + "\n"))
					username = client.username
					continue
				}
				client.username = ""
				continue
			}
			server.Broadcast(message, username)
		case message := <-client.outbound:
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
				return
			}
			if n > 1 {
				incoming <- strings.TrimSuffix(string(buf[:n]), "\r\n")
			}
		}
	}()
	return
}
