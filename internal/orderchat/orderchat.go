package orderchat

import (
	"sync"

	"github.com/gorilla/websocket"

	"ptop/internal/models"
)

var clients = struct {
	sync.RWMutex
	m map[string]map[*websocket.Conn]bool
}{m: make(map[string]map[*websocket.Conn]bool)}

func AddClient(chatID string, conn *websocket.Conn) {
	clients.Lock()
	defer clients.Unlock()
	conns, ok := clients.m[chatID]
	if !ok {
		conns = make(map[*websocket.Conn]bool)
		clients.m[chatID] = conns
	}
	conns[conn] = true
}

func RemoveClient(chatID string, conn *websocket.Conn) {
	clients.Lock()
	defer clients.Unlock()
	if conns, ok := clients.m[chatID]; ok {
		delete(conns, conn)
	}
}

func Send(conn *websocket.Conn, msg models.OrderMessage) error {
	return conn.WriteJSON(msg)
}

func Broadcast(chatID string, msg models.OrderMessage) {
	clients.Lock()
	defer clients.Unlock()
	for c := range clients.m[chatID] {
		if err := Send(c, msg); err != nil {
			c.Close()
			delete(clients.m[chatID], c)
		}
	}
}
