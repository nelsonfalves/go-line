package model

import "net"

type Message struct {
	Sender  string
	Content []byte
}

type Sender struct {
	Conn     net.Conn
	Username string
}

func (m Message) Bytes() []byte {
	return append([]byte(m.Sender+": "), m.Content...)
}

type Room struct {
	Name     string
	Password string
}
