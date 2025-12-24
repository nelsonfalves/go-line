package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nelsonfalves/go-line/internal/client"
	"github.com/nelsonfalves/go-line/internal/server"
)

const (
	protocol = "tcp"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Println("Must provide at least three arguments.")
		return
	}

	port := ":" + args[1]
	mode := args[2]
	roomName := args[3]
	username := args[4]
	password := args[5]

	if mode == "server" {
		s := server.New(roomName, password)
		go s.Start(port)

		time.Sleep(100 * time.Millisecond)

		c := client.New(username, password)
		c.Connect(port)
	}

	if mode == "client" {
		c := client.New(username, password)
		c.Connect(port)
	}
}
