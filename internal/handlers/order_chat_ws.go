package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/services"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var orderChatClients = struct {
	sync.RWMutex
	m map[string]map[*websocket.Conn]bool
}{m: make(map[string]map[*websocket.Conn]bool)}

type orderChatEvent struct {
	Type    string              `json:"type"`
	Message models.OrderMessage `json:"message"`
}

func newOrderChatEvent(msg models.OrderMessage) orderChatEvent {
	return orderChatEvent{Type: string(msg.Type), Message: msg}
}

func sendOrderChatEvent(conn *websocket.Conn, msg models.OrderMessage) error {
	return conn.WriteJSON(newOrderChatEvent(msg))
}

func broadcastOrderChatMessage(chatID string, msg models.OrderMessage) {
	orderChatClients.Lock()
	for c := range orderChatClients.m[chatID] {
		if err := sendOrderChatEvent(c, msg); err != nil {
			c.Close()
			delete(orderChatClients.m[chatID], c)
		}
	}
	orderChatClients.Unlock()
}

// OrderChatWS godoc
// @Summary Websocket чат ордера
// @Description При подключении отправляет историю сообщений из кеша Redis
// @Tags orders
// @Security BearerAuth
// @Param id path string true "ID ордера"
// @Success 101 {string} string "Switching Protocols"
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /ws/orders/{id}/chat [get]
func OrderChatWS(db *gorm.DB, cache *services.ChatCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var order models.Order
		if err := db.Where("id = ?", orderID).First(&order).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid order"})
			return
		}
		if clientID != order.BuyerID && clientID != order.SellerID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		var chat models.OrderChat
		if err := db.Where("order_id = ?", order.ID).First(&chat).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				chat = models.OrderChat{OrderID: order.ID}
				if err := db.Create(&chat).Error; err != nil {
					c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
				return
			}
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		orderChatClients.Lock()
		conns, ok := orderChatClients.m[chat.ID]
		if !ok {
			conns = make(map[*websocket.Conn]bool)
			orderChatClients.m[chat.ID] = conns
		}
		conns[conn] = true
		orderChatClients.Unlock()
		defer func() {
			orderChatClients.Lock()
			delete(orderChatClients.m[chat.ID], conn)
			orderChatClients.Unlock()
		}()

		if cache != nil {
			if history, err := cache.GetHistory(c.Request.Context(), chat.ID); err == nil {
				for _, m := range history {
					if err := sendOrderChatEvent(conn, m); err != nil {
						return
					}
				}
			}
		}

		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var r OrderMessageRequest
			if err := json.Unmarshal(msgBytes, &r); err != nil || r.Content == "" {
				continue
			}
			msg := models.OrderMessage{ChatID: chat.ID, ClientID: clientID, Type: models.MessageTypeText, Content: r.Content}
			if err := db.Create(&msg).Error; err != nil {
				continue
			}
			if cache != nil {
				_ = cache.AddMessage(c.Request.Context(), chat.ID, msg)
			}
			broadcastOrderChatMessage(chat.ID, msg)
		}
	}
}
