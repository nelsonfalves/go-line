package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type Client interface {
	Connect(port string)
}

type client struct{}

func New() Client {
	return &client{}
}

func (c *client) Connect(port string) {
	addr := "localhost" + port
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	c.read(conn)
}

func (c *client) read(conn net.Conn) {
	go func() {
		buffer := make([]byte, 4096)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Disconnected from server")
				return
			}
			msg := buffer[:n]

			fmt.Print(string(msg))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		_, err := conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("Error sending:", err)
			return
		}
	}
}
