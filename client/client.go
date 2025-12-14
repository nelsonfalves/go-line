package client

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/nelsonfalves/go-line/constant"
)

type Client interface {
	Connect(port string)
}

type client struct {
	name string
}

func New(name string) Client {
	return &client{
		name: name,
	}
}

func (c *client) Connect(port string) {
	addr := "localhost" + port
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte(c.name))

	c.read(conn)
}

func (c *client) read(conn net.Conn) {
	go func() {
		buffer := make([]byte, constant.DefaultBufferSize)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Disconnected from server")
				return
			}
			content := buffer[:n]

			fmt.Print(string(content))
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
		// Clear the typed message from terminal (move cursor up one line and clear it)
		fmt.Print("\033[1A\033[2K")
	}
}
