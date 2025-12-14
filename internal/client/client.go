package client

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/nelsonfalves/go-line/internal/constant"
)

type Client interface {
	Connect(port string)
}

type client struct {
	name string
	conn net.Conn
}

func New(name string) Client {
	return &client{
		name: name,
	}
}

func (c *client) Connect(port string) {
	conn, err := net.Dial(constant.DefaultProtocol, constant.DefaultHost+port)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	c.conn = conn
	conn.Write([]byte(c.name))

	c.startCommunication()
}

func (c *client) startCommunication() {
	go c.receiveMessages()
	c.sendMessages()
}

func (c *client) receiveMessages() {
	buffer := make([]byte, constant.DefaultBufferSize)
	for {
		n, err := c.conn.Read(buffer)
		if err != nil {
			fmt.Println("Disconnected from server")
			return
		}
		content := buffer[:n]

		fmt.Print(string(content))
	}
}

func (c *client) sendMessages() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		_, err := c.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("Error sending:", err)
			return
		}
		// Clear the typed message from terminal (move cursor up one line and clear it)
		fmt.Print(constant.ClearTypedMessage)
	}
}
