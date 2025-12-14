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
	room  *model.Room
	mutex sync.RWMutex
}

func New(name, password string) Server {
	return &server{
		room: &model.Room{
			Name:     name,
			Password: password,
			Clients:  make(map[net.Conn]string),
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
		conn.Write([]byte(fmt.Sprintf("ERROR: %s\n", err.Error())))
		return
	}

	if password != s.room.Password {
		conn.Write([]byte("ERROR: Wrong password\n"))
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
			Content: content,
			Sender:  username,
		}

		s.broadcastMessage(msg, conn)
	}
}

func (s *server) register(conn net.Conn, name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.room.Clients) == 0 {
		s.room.Host = conn
		fmt.Printf("Host '%s' created room '%s'\n", name, s.room.Name)
	}

	s.room.Clients[conn] = name
	fmt.Printf("Client connected: %s (Total clients: %d)\n", name, len(s.room.Clients))
}

func (s *server) broadcastMessage(msg model.Message, sender net.Conn) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clients := s.room.Clients

	for conn := range clients {
		if conn == sender {
			continue
		}

		_, err := conn.Write(msg.Bytes())
		if err != nil {
			clientName := clients[conn]
			fmt.Printf("Error sending message to %s: %v\n", clientName, err)
			continue
		}
	}
}

func (s *server) close(conn net.Conn) {
	defer conn.Close()

	clients := s.room.Clients
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
