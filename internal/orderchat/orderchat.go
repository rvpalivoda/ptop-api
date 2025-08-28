package orderchat

import (
	"sync"

	"github.com/gorilla/websocket"

	"ptop/internal/models"
)

type Event struct {
    Type    string              `json:"type"`
    Message models.OrderMessage `json:"message"`
}

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

func newEvent(msg models.OrderMessage) Event {
	return Event{Type: string(msg.Type), Message: msg}
}

func Send(conn *websocket.Conn, msg models.OrderMessage) error {
	return conn.WriteJSON(newEvent(msg))
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

// BroadcastRead рассылает событие о прочтении сообщения
func BroadcastRead(chatID string, msg models.OrderMessage) {
    clients.Lock()
    defer clients.Unlock()
    evt := Event{Type: "READ", Message: msg}
    for c := range clients.m[chatID] {
        if err := c.WriteJSON(evt); err != nil {
            c.Close()
            delete(clients.m[chatID], c)
        }
    }
}
