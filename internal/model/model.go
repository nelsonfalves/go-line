package model

type Message struct {
	Sender  string
	Content []byte
}

func (m Message) Bytes() []byte {
	return append([]byte(m.Sender+": "), m.Content...)
}
