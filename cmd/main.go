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
	if len(args) < 2 {
		fmt.Println("Must provide at least two arguments.")
		return
	}

	port := ":" + args[1]

	if args[2] == "server" {
		s := server.New()
		s.Start(port)
	}

	if args[2] == "client" {
		name := args[3]
		c := client.New(name)
		c.Connect(port)
	}

}
