package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{}

type Message struct {
	Type string          `json:"type"`
	UUID string          `json:"uuid"`
	Data json.RawMessage `json:"data"`
}

type LocationData struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during upgrade:", err)
		return
	}

	for {
		var msg Message
		if err := c.ReadJSON(&msg); err != nil {
			log.Println("Error during read:", err)
			break
		}

		uuid := msg.UUID

		if uuid == "" {
			log.Println("No UUID provided")
			break
		}

		switch msg.Type {
		case "connect":
			handleConnect(c, uuid)
			break
		case "disconnect":
			handleDisconnect(c, uuid)
			break
		case "location":
			var data LocationData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Println("Error during unmarshal:", err)
				break
			}

			handleLocation(c, uuid, data)
		default:
			log.Println("Unknown message type:", msg.Type)
		}
	}

	c.Close()
}

func handleConnect(conn *websocket.Conn, uuid string) {
	clients.Lock()
	defer clients.Unlock()

	if c, ok := clients.m[uuid]; ok {
		c.Conn = conn
		c.Connected = true
		c.LastSeen = time.Now().Format(time.RFC1123)
		clients.m[uuid] = c
	} else {
		clients.m[uuid] = Client{
			Conn:      conn,
			UUID:      uuid,
			Connected: true,
			LastSeen:  time.Now().Format(time.RFC1123),
		}
	}

	log.Println("Client connected:", uuid)
}

func handleDisconnect(conn *websocket.Conn, uuid string) {
	clients.Lock()
	defer clients.Unlock()

	if c, ok := clients.m[uuid]; ok {
		c.Connected = false
		c.LastSeen = time.Now().Format(time.RFC1123)
		clients.m[uuid] = c
		log.Println("Client disconnected:", uuid)
	} else {
		log.Println("Client not found:", uuid)
	}
}

func handleLocation(conn *websocket.Conn, uuid string, data LocationData) {
	clients.Lock()
	defer clients.Unlock()

	if c, ok := clients.m[uuid]; ok {
		c.Location = data
		c.LastSeen = time.Now().Format(time.RFC1123)
		clients.m[uuid] = c
		log.Println("Location updated:", uuid)
	} else {
		log.Println("Client not found:", uuid)
	}
}
