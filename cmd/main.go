package main

import (
	"fmt"
	"os"

	"github.com/nelsonfalves/go-line/client"
	"github.com/nelsonfalves/go-line/server"
)

const (
	protocol = "tcp"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Println("Please provide the initiation mode and a port number!")
		return
	}

	port := ":" + args[2]

	if args[1] == "server" {
		s := server.New()
		s.Start(port)
	}

	if args[1] == "client" {
		c := client.New()
		c.Connect(port)
	}

}
