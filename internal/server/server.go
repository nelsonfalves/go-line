package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/nelsonfalves/go-line/internal/constant"
	"github.com/nelsonfalves/go-line/internal/model"
)

type Server interface {
	Start(port string)
}

type server struct {
	room    *model.Room
	host    net.Conn
	mutex   sync.RWMutex
	clients map[net.Conn]string
}

func New(name, password string) Server {
	return &server{
		room: &model.Room{
			Name:     name,
			Password: password,
		},
	}
}

func (s *server) Start(port string) {
	listener, err := net.Listen(constant.DefaultProtocol, port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("Server listening on %s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go s.handleClient(conn)
	}
}

func (s *server) handleClient(conn net.Conn) {
	defer s.close(conn)

	buffer := make([]byte, constant.DefaultBufferSize)
	n, err := conn.Read(buffer)
	if err != nil {
		return
	}

	username, password, err := extractCredentials(buffer, n)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("error: %s\n", err.Error())))
		return
	}

	if password != s.room.Password {
		conn.Write([]byte("error: wrong password\n"))
		return
	}

	s.register(conn, username)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			return
		}

		content := buffer[:n]
		msg := model.Message{
			Sender:  username,
			Content: content,
		}

		s.broadcast(conn, msg)
	}
}

func (s *server) register(conn net.Conn, username string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.clients) == 0 {
		s.host = conn
		fmt.Printf("Host '%s' created room '%s'\n", username, s.room.Name)
	}

	s.clients[conn] = username
	fmt.Printf("Client connected: %s (Total clients: %d)\n", username, len(s.clients))
}

func (s *server) broadcast(sender net.Conn, msg model.Message) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clients := s.clients

	for client := range clients {
		if client == sender {
			continue
		}

		_, err := client.Write(msg.Bytes())
		if err != nil {
			clientName := clients[client]
			fmt.Printf("Error sending message to %s: %v\n", clientName, err)
			continue
		}
	}
}

func (s *server) close(conn net.Conn) {
	defer conn.Close()

	clients := s.clients
	clientName := clients[conn]

	s.mutex.Lock()
	delete(clients, conn)
	fmt.Printf("Client disconnected: %s (Total clients: %d)\n", clientName, len(clients))
	s.mutex.Unlock()
}

func extractCredentials(buffer []byte, n int) (string, string, error) {
	parts := strings.Split(string(buffer[:n]), "\n")
	if len(parts) != 2 {
		return "", "", errors.New("invalid credentials format")
	}

	username := strings.TrimSpace(parts[0])
	password := strings.TrimSpace(parts[1])

	if username == "" {
		return "", "", errors.New("username cannot be empty")
	}

	if password == "" {
		return "", "", errors.New("password cannot be empty")
	}

	return username, password, nil
}
