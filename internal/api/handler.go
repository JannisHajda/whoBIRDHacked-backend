package api

import (
	"encoding/json"
	"github.com/JannisHajda/whoBIRDHacked-backend/internal/ws"
	"net/http"
)

type Message struct {
	Cmd  string `json:"cmd"`
	UUID string `json:"uuid"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if msg.UUID == "" {
		http.Error(w, "Missing UUID", http.StatusBadRequest)
		return
	}

	cmd := msg.Cmd
	uuid := msg.UUID

	client, ok := ws.GetClient(uuid)
	if !ok {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	switch cmd {
	case "ping":
		err := client.Ping()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Unknown command", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Command executed"))
}
