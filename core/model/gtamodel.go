package model

type EventItem struct {
	Type     int
	Conn     interface{}
	UserData interface{}
}

type Config struct {
	Id   int
	Host string
	Port int
}
