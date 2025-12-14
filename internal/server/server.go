package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/nelsonfalves/go-line/internal/constant"
	"github.com/nelsonfalves/go-line/internal/model"
)

type Server interface {
	Start(port string)
}

type server struct {
	clients map[net.Conn]string
	mutex   sync.RWMutex
}

func New() Server {
	return &server{
		clients: make(map[net.Conn]string),
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

	name := string(buffer[:n])
	s.register(conn, name)

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
			Sender:  name,
		}

		s.broadcastMessage(msg, conn)
	}
}

func (s *server) register(conn net.Conn, name string) {
	s.mutex.Lock()
	s.clients[conn] = name
	fmt.Printf("Client connected: %s (Total clients: %d)\n", name, len(s.clients))
	s.mutex.Unlock()
}

func (s *server) broadcastMessage(msg model.Message, sender net.Conn) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for conn := range s.clients {
		if conn == sender {
			continue
		}

		_, err := conn.Write(msg.Bytes())
		if err != nil {
			addr := conn.RemoteAddr().String()
			fmt.Printf("Error broadcasting to %s: %v\n", addr, err)
			continue
		}
	}
}

func (s *server) close(conn net.Conn) {
	s.mutex.Lock()
	delete(s.clients, conn)
	fmt.Printf("Client disconnected: %s (Total clients: %d)\n", conn.RemoteAddr().String(), len(s.clients))
	s.mutex.Unlock()
	conn.Close()
}
