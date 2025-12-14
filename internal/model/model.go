package model

import "net"

type Message struct {
	Sender  string
	Content []byte
}

func (m Message) Bytes() []byte {
	return append([]byte(m.Sender+": "), m.Content...)
}

type Room struct {
	Name     string
	Password string
	Host     net.Conn
	Clients  map[net.Conn]string
}
