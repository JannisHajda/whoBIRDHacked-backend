package ws

import "sync"

var clients = struct {
	sync.RWMutex
	m map[string]Client
}{
	m: make(map[string]Client),
}

func GetClients() map[string]Client {
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
