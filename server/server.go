package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/nelsonfalves/go-line/constant"
	"github.com/nelsonfalves/go-line/model"
)

const (
	protocol = "tcp"
)

type Server interface {
	Start(port string)
}

type server struct {
	clients map[net.Conn]model.User
	mutex   sync.RWMutex
}

func New() Server {
	return &server{
		clients: make(map[net.Conn]model.User),
	}
}

func (s *server) Start(port string) {
	listener, err := net.Listen(protocol, port)
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

		go s.handle(conn)
	}
}

func (s *server) handle(conn net.Conn) {
	defer s.delete(conn)

	buffer := make([]byte, constant.DefaultBufferSize)

	n, err := conn.Read(buffer)
	if err != nil {
		return
	}

	name := string(buffer[:n])
	user := model.NewUser(name)

	s.register(conn, user)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			return
		}

		content := buffer[:n]
		msg := model.NewMessage(user.Name, content)

		s.broadcast(msg)
	}
}

func (s *server) register(conn net.Conn, user model.User) {
	s.mutex.Lock()
	s.clients[conn] = user
	fmt.Printf("Client connected: %s (Total clients: %d)\n", user.Name, len(s.clients))
	s.mutex.Unlock()
}

func (s *server) broadcast(msg model.Message) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for client := range s.clients {
		addr := client.RemoteAddr().String()
		if addr == msg.Sender {
			continue
		}

		_, err := client.Write(msg.Bytes())
		if err != nil {
			fmt.Printf("Error broadcasting to %s: %v\n", addr, err)
			continue
		}
	}
}

func (s *server) delete(conn net.Conn) {
	s.mutex.Lock()
	delete(s.clients, conn)
	fmt.Printf("Client disconnected: %s (Total clients: %d)\n", conn.RemoteAddr().String(), len(s.clients))
	s.mutex.Unlock()
	conn.Close()
}
