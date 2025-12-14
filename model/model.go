package model

type Room struct {
	Name     string
	Password string
	Clients  map[string]*User
}

type User struct {
	Name string
}

func NewUser(name string) User {
	return User{
		Name: name,
	}
}

type Message struct {
	Sender  string
	Content []byte
}

func NewMessage(sender string, content []byte) Message {
	return Message{
		Sender:  sender,
		Content: content,
	}
}

func (m Message) Bytes() []byte {
	return append([]byte(m.Sender+": "), m.Content...)
}
