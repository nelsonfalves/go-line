package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type Server struct {
	conn net.Conn
}

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	port := ":" + args[1]
	listener, err := net.Listen("tcp4", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		server := Server{conn: c}
		go server.handleConnection()
	}
}

func (s Server) handleConnection() {
	defer s.conn.Close()

	tmp := make([]byte, 4096)
	for {
		n, err := s.conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}
		msg := tmp[:n]

		response := fmt.Sprintf("Server received: %s", string(msg))
		_, err = s.conn.Write([]byte(response))
		if err != nil {
			fmt.Println("write error: ", err)
		}
	}
}
