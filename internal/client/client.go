package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/nelsonfalves/go-line/internal/constant"
)

type Client interface {
	Connect(port string) error
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

func (c *client) Connect(port string) error {
	addr := constant.DefaultHost + port
	conn, err := net.Dial(constant.DefaultProtocol, addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	c.conn = conn

	if _, err := conn.Write([]byte(c.username + "\n" + c.password + "\n")); err != nil {
		return fmt.Errorf("failed to send credentials: %w", err)
	}

	if err := c.waitForServerResponse(); err != nil {
		return err
	}

	fmt.Printf("connected to %s\n", addr)

	return c.startCommunication()
}

func (c *client) waitForServerResponse() error {
	buffer := make([]byte, constant.DefaultBufferSize)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to receive server response: %w", err)
	}

	content := buffer[:n]
	response := strings.TrimSpace(string(content))
	if strings.HasPrefix(response, constant.ErrorPrefix) {
		errorMsg := strings.TrimSpace(strings.TrimPrefix(response, constant.ErrorPrefix))
		return fmt.Errorf("server rejected connection: %s", errorMsg)
	}

	if response != "OK" {
		return fmt.Errorf("unexpected server response: %s", response)
	}

	return nil
}

func (c *client) startCommunication() error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- c.receiveMessages()
	}()

	go func() {
		errChan <- c.sendMessages()
	}()

	return <-errChan
}

func (c *client) receiveMessages() error {
	buffer := make([]byte, constant.DefaultBufferSize)
	for {
		n, err := c.conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("server closed connection: %w", err)
			}
			return fmt.Errorf("failed to read from server: %w", err)
		}
		content := buffer[:n]

		fmt.Print(string(content))
	}
}

func (c *client) sendMessages() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		if _, err := c.conn.Write([]byte(msg + "\n")); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		// Clear the typed message from terminal (move cursor up one line and clear it)
		fmt.Print(constant.ClearTypedMessage)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}
	return fmt.Errorf("stdin closed: EOF")
}
