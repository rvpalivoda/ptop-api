package notifications

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"ptop/internal/models"
)

var (
	db      *gorm.DB
	clients = struct {
		sync.RWMutex
		m map[string]map[*websocket.Conn]bool
	}{m: make(map[string]map[*websocket.Conn]bool)}
)

// SetDB устанавливает соединение с базой данных для обновления уведомлений.
func SetDB(d *gorm.DB) {
	db = d
}

// AddClient добавляет соединение вебсокета для клиента.
func AddClient(clientID string, conn *websocket.Conn) {
	clients.Lock()
	defer clients.Unlock()
	conns, ok := clients.m[clientID]
	if !ok {
		conns = make(map[*websocket.Conn]bool)
		clients.m[clientID] = conns
	}
	conns[conn] = true
}

// RemoveClient удаляет соединение вебсокета для клиента.
func RemoveClient(clientID string, conn *websocket.Conn) {
	clients.Lock()
	defer clients.Unlock()
	if conns, ok := clients.m[clientID]; ok {
		delete(conns, conn)
	}
}

// Send отправляет уведомление через указанное соединение.
// При успешной отправке поле SentAt обновляется в базе данных.
func Send(conn *websocket.Conn, n models.Notification) error {
	if err := conn.WriteJSON(n); err != nil {
		return err
	}
	if db != nil {
		now := time.Now()
		db.Model(&models.Notification{}).Where("id = ?", n.ID).Update("sent_at", now)
	}
	return nil
}

// Broadcast отправляет уведомление всем соединениям клиента.
func Broadcast(clientID string, n models.Notification) {
	clients.Lock()
	defer clients.Unlock()
	for c := range clients.m[clientID] {
		if err := Send(c, n); err != nil {
			c.Close()
			delete(clients.m[clientID], c)
		}
	}
}
