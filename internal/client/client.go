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
	conn     net.Conn
	username string
	password string
}

func New(username, password string) Client {
	return &client{
		username: username,
		password: password,
	}
}

func (c *client) Connect(port string) {
	conn, err := net.Dial(constant.DefaultProtocol, constant.DefaultHost+port)
	if err != nil {
		fmt.Println("error connecting:", err)
		return
	}
	defer conn.Close()

	c.conn = conn
	conn.Write([]byte(c.username + "\n" + c.password + "\n"))

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
			fmt.Println("disconnected from server")
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
			fmt.Println("error sending:", err)
			return
		}
		// Clear the typed message from terminal (move cursor up one line and clear it)
		fmt.Print(constant.ClearTypedMessage)
	}
}
