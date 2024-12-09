package ws

import (
	"fmt"
	"sync"
	"time"
)

var clients = struct {
	sync.RWMutex
	m map[string]Client
}{
	m: make(map[string]Client),
}

var timeout = 30 * time.Second

func filterInactive() {
	clients.Lock()
	defer clients.Unlock()

	for _, client := range clients.m {
		lastSeen, err := time.Parse(time.RFC1123, client.LastSeen)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			continue
		}

		if client.Connected && time.Since(lastSeen) > timeout {
			client.Connected = false
			clients.m[client.UUID] = client
		}
	}
}

func GetClients() map[string]Client {
	filterInactive()
	clients.RLock()
	defer clients.RUnlock()
	return clients.m
}

func GetClient(uuid string) (Client, bool) {
	clients.RLock()
	defer clients.RUnlock()

	client, ok := clients.m[uuid]
	return client, ok
}
