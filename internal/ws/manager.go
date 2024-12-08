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

	for _, c := range clients.m {
		lastSeen, err := time.Parse(time.RFC1123, c.LastSeen)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			continue
		}

		if c.Connected && time.Since(lastSeen) > timeout {
			c := clients.m[c.UUID]
			c.Connected = false
			clients.m[c.UUID] = c
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

	c, ok := clients.m[uuid]
	return c, ok
}
