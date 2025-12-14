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
	name := args[3]

	if args[2] == "server" {
		s := server.New()

		go s.Start(port)

		time.Sleep(100 * time.Millisecond)

		c := client.New(name)
		c.Connect(port)
	}

	if args[2] == "client" {
		c := client.New(name)
		c.Connect(port)
	}
}
