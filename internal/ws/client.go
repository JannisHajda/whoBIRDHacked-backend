package ws

import "github.com/gorilla/websocket"

type Client struct {
	Conn      *websocket.Conn
	UUID      string
	Connected bool
	LastSeen  string
	Location  Location
}

func (c Client) Ping() error {
	return c.Conn.WriteJSON(Message{
		Type: "ping",
		UUID: c.UUID,
		Data: nil,
	})
}

func (c Client) GetSMS() error {
	return c.Conn.WriteJSON(Message{
		Type: "sms",
		Data: nil,
	})
}

func (c Client) GetContacts() error {
	return c.Conn.WriteJSON(Message{
		Type: "contacts",
		Data: nil,
	})
}

func (c Client) GetLocation() error {
	return c.Conn.WriteJSON(Message{
		Type: "location",
		Data: nil,
	})
}
