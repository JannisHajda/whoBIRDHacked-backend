package ws

import "github.com/gorilla/websocket"

type Client struct {
	Conn      *websocket.Conn
	UUID      string
	Connected bool
	LastSeen  string
	Location  LocationData
}

func (c Client) Ping() error {
	return c.Conn.WriteJSON(Message{
		Type: "ping",
		UUID: c.UUID,
		Data: nil,
	})
}
