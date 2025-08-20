package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"ptop/internal/models"
)

var orderClients = struct {
	sync.RWMutex
	m map[string]map[*websocket.Conn]bool
}{m: make(map[string]map[*websocket.Conn]bool)}

// OrderEvent событие, отправляемое клиенту при создании его ордера.
// Type всегда имеет значение `order.created`.
type OrderEvent struct {
	Type  string           `json:"type" example:"order.created"`
	Order models.OrderFull `json:"order"`
}

func newOrderEvent(of models.OrderFull) OrderEvent {
	return OrderEvent{Type: "order.created", Order: of}
}

func broadcastOrderEvent(clientID string, evt OrderEvent) {
	orderClients.Lock()
	defer orderClients.Unlock()
	for c := range orderClients.m[clientID] {
		if err := c.WriteJSON(evt); err != nil {
			c.Close()
			delete(orderClients.m[clientID], c)
		}
	}
}

// OrdersWS godoc
// @Summary Websocket ордеров клиента
// @Description После подключения авторизованный клиент получает события OrderEvent о создании своих ордеров.
// Клиенту не нужно отправлять данные, соединение используется только для чтения.
// @Tags orders
// @Security BearerAuth
// @Success 101 {object} handlers.OrderEvent "Switching Protocols"
// @Failure 401 {object} ErrorResponse
// @Router /ws/orders [get]
func OrdersWS() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		orderClients.Lock()
		conns, ok := orderClients.m[clientID]
		if !ok {
			conns = make(map[*websocket.Conn]bool)
			orderClients.m[clientID] = conns
		}
		conns[conn] = true
		orderClients.Unlock()
		defer func() {
			orderClients.Lock()
			delete(orderClients.m[clientID], conn)
			orderClients.Unlock()
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}
