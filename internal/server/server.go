package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/nelsonfalves/go-line/internal/constant"
	"github.com/nelsonfalves/go-line/internal/model"
)

type Server interface {
	Start(port string) error
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
		clients: make(map[net.Conn]string),
	}
}

func (s *server) Start(port string) error {
	listener, err := net.Listen(constant.DefaultProtocol, port)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer listener.Close()

	fmt.Printf("server listening on %s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		go func() {
			if err := s.handleClient(conn); err != nil {
				fmt.Printf("client error: %v\n", err)
			}
		}()
	}
}

func (s *server) handleClient(conn net.Conn) error {
	defer s.close(conn)

	buffer := make([]byte, constant.DefaultBufferSize)
	n, err := conn.Read(buffer)
	if err != nil {
		connErr := fmt.Errorf("failed to read credentials: %w", err)
		s.sendError(conn, connErr)
		return connErr
	}

	content := buffer[:n]
	username, password, err := extractCredentials(content)
	if err != nil {
		connErr := fmt.Errorf("invalid credentials: %w", err)
		s.sendError(conn, connErr)
		return connErr
	}

	if err := s.authenticate(password); err != nil {
		connErr := fmt.Errorf("authentication failed: %w", err)
		s.sendError(conn, connErr)
		return connErr
	}

	s.register(conn, username)

	if _, err := conn.Write([]byte("OK\n")); err != nil {
		return fmt.Errorf("failed to send success response: %w", err)
	}

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read from client: %w", err)
		}

		content := buffer[:n]
		msg := model.Message{
			Sender:  username,
			Content: content,
		}

		s.broadcast(msg)
	}
}

func (s *server) sendError(conn net.Conn, connErr error) {
	errorMsg := fmt.Sprintf("error: %v\n", connErr)

	if _, err := conn.Write([]byte(errorMsg)); err != nil {
		fmt.Printf("failed to send error to client: %v\n", err)
	}
}

func (s *server) register(conn net.Conn, username string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.clients) == 0 {
		s.host = conn
		fmt.Printf("host '%s' created room '%s'\n", username, s.room.Name)
	}

	s.clients[conn] = username
	fmt.Printf("client connected: %s (total clients: %d)\n", username, len(s.clients))
}

func (s *server) broadcast(msg model.Message) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clients := s.clients

	for client := range clients {
		_, err := client.Write(msg.Bytes())
		if err != nil {
			clientName := clients[client]
			fmt.Printf("failed to send message to %s: %v\n", clientName, err)
			continue
		}
	}
}

func (s *server) close(conn net.Conn) {
	defer conn.Close()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if client, exists := s.clients[conn]; exists {
		delete(s.clients, conn)
		fmt.Printf("client disconnected: %s (total remaining clients: %d)\n", client, len(s.clients))
		return
	}

	fmt.Printf("connection closed before registration (total clients: %d)\n", len(s.clients))
}

func extractCredentials(content []byte) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(string(content)), "\n", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid credentials")
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

func (s *server) authenticate(password string) error {
	if s.room.Password != password {
		return errors.New("incorrect password")
	}
	return nil
}
