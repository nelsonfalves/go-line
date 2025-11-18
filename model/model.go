package model

type Room struct {
	Name     string
	Password string
	Clients  map[string]*Client
}

type Client struct {
	Name string
}
