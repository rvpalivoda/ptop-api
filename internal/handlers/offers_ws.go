package handlers

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"ptop/internal/models"
)

// offerEvent описывает событие оффера, передаваемое по websocket.
type offerEvent struct {
	Type  string           `json:"type"`
	Offer models.OfferFull `json:"offer"`
}

// канал -> множество подключений
var offerWSConns = struct {
	sync.Mutex
	m map[string]map[*websocket.Conn]bool
}{m: make(map[string]map[*websocket.Conn]bool)}

// OffersWS godoc
// @Summary WebSocket обновления офферов
// @Description Подключение для получения событий создания, обновления и удаления объявлений
// @Tags offers
// @Security BearerAuth
// @Param channel query string false "канал"
// @Success 101 {string} string "Switching Protocols"
// @Failure 403 {object} ErrorResponse
// @Router /ws/offers [get]
func OffersWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := r.URL.Query().Get("channel")
		if channel == "" {
			channel = "offers"
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// опционально: http.Error(w, "upgrade failed", http.StatusBadRequest)
			return
		}

		// Регистрируем подключение
		offerWSConns.Lock()
		conns, ok := offerWSConns.m[channel]
		if !ok {
			conns = make(map[*websocket.Conn]bool)
			offerWSConns.m[channel] = conns
		}
		conns[conn] = true
		offerWSConns.Unlock()

		// Удаляем подключение при выходе
		defer func() {
			offerWSConns.Lock()
			if conns, ok := offerWSConns.m[channel]; ok {
				delete(conns, conn)
				if len(conns) == 0 {
					delete(offerWSConns.m, channel)
				}
			}
			offerWSConns.Unlock()
			conn.Close()
		}()

		// Держим соединение открытым, читаем любые входящие сообщения (ping/pong/json)
		for {
			var v interface{}
			if err := conn.ReadJSON(&v); err != nil {
				break
			}
		}
	}
}

func broadcastOfferEvent(eventType string, offer models.OfferFull) {
	const channel = "offers"

	// Делаем снимок подключений, чтобы не держать мьютекс во время записи
	offerWSConns.Lock()
	connsMap := offerWSConns.m[channel]
	snapshot := make([]*websocket.Conn, 0, len(connsMap))
	for c := range connsMap {
		snapshot = append(snapshot, c)
	}
	offerWSConns.Unlock()

	// Пишем события; проблемные соединения отписываем
	for _, c := range snapshot {
		if err := c.WriteJSON(offerEvent{Type: eventType, Offer: offer}); err != nil {
			c.Close()
			offerWSConns.Lock()
			if conns, ok := offerWSConns.m[channel]; ok {
				delete(conns, c)
				if len(conns) == 0 {
					delete(offerWSConns.m, channel)
				}
			}
			offerWSConns.Unlock()
		}
	}
}
