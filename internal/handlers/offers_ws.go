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

var offerWSConns = struct {
	sync.Mutex
	m map[string][]*websocket.Conn
}{m: make(map[string][]*websocket.Conn)}

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
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		offerWSConns.Lock()
		offerWSConns.m[channel] = append(offerWSConns.m[channel], ws)
		offerWSConns.Unlock()

		defer func() {
			offerWSConns.Lock()
			conns := offerWSConns.m[channel]
			for i, c := range conns {
				if c == ws {
					offerWSConns.m[channel] = append(conns[:i], conns[i+1:]...)
					break
				}
			}
			offerWSConns.Unlock()
			ws.Close()
		}()

		for {
			var v interface{}
			if err := ws.ReadJSON(&v); err != nil {
				break
			}
		}
	}
}

func broadcastOfferEvent(eventType string, offer models.OfferFull) {
	channel := "offers"
	offerWSConns.Lock()
	conns := offerWSConns.m[channel]
	for i := 0; i < len(conns); {
		if err := conns[i].WriteJSON(offerEvent{Type: eventType, Offer: offer}); err != nil {
			conns[i].Close()
			conns = append(conns[:i], conns[i+1:]...)
		} else {
			i++
		}
	}
	offerWSConns.m[channel] = conns
	offerWSConns.Unlock()
}
