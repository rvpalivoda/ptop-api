package handlers

import (
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
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
		websocket.Handler(func(ws *websocket.Conn) {
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
				if err := websocket.JSON.Receive(ws, &v); err != nil {
					break
				}
			}
		}).ServeHTTP(w, r)
	}
}

func broadcastOfferEvent(eventType string, offer models.OfferFull) {
	channel := "offers"
	offerWSConns.Lock()
	conns := offerWSConns.m[channel]
	for i := 0; i < len(conns); {
		if err := websocket.JSON.Send(conns[i], offerEvent{Type: eventType, Offer: offer}); err != nil {
			conns[i].Close()
			conns = append(conns[:i], conns[i+1:]...)
		} else {
			i++
		}
	}
	offerWSConns.m[channel] = conns
	offerWSConns.Unlock()
}
