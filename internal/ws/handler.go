package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/exec"
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

type AudioData []int16

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
		case "audio":
			var data AudioData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Println("Error during unmarshal:", err)
				break
			}
			handleAudio(c, uuid, data)
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

func handleAudio(conn *websocket.Conn, uuid string, data AudioData) {
	// Data is already unmarshalled as AudioData []int16
	log.Printf("Received audio data of length: %d\n", len(data))

	// Convert the data to raw bytes
	audioBytes := int16ToBytes(data)

	// Save raw audio data as a file
	rawFileName := fmt.Sprintf("%s_audio.raw", uuid)
	if err := saveRawAudio(rawFileName, audioBytes); err != nil {
		log.Println("Error saving raw audio:", err)
		return
	}

	// Convert raw audio to MP3 using FFmpeg
	mp3FileName := fmt.Sprintf("%s_audio.mp3", uuid)
	if err := convertToMP3(rawFileName, mp3FileName); err != nil {
		log.Println("Error converting to MP3:", err)
		return
	}

	log.Printf("Saved MP3 audio for UUID %s at %s\n", uuid, mp3FileName)
}

func int16ToBytes(data []int16) []byte {
	byteData := make([]byte, len(data)*2)
	for i, v := range data {
		byteData[i*2] = byte(v)
		byteData[i*2+1] = byte(v >> 8)
	}
	return byteData
}

func saveRawAudio(fileName string, data []byte) error {
	return os.WriteFile(fileName, data, 0644)
}

func convertToMP3(inputFile, outputFile string) error {
	cmd := exec.Command("ffmpeg",
		"-f", "s16le", // Raw audio format
		"-ar", "48000", // Sample rate (adjust as needed)
		"-ac", "1", // Number of channels (mono)
		"-i", inputFile, // Input file
		outputFile, // Output file
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %s\nOutput: %s", err, string(output))
	}

	return nil
}
