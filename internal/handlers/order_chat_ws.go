package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/notifications"
	"ptop/internal/orderchat"
	"ptop/internal/services"
)

// OrderChatWS godoc
// @Summary Websocket чат ордера
// @Description Подключает покупателя и продавца к чату ордера.
// После подключения сервер отправляет историю сообщений (models.OrderMessage).
// Каждое сообщение содержит поле senderName — username отправителя (по client_id).
// Сервер также рассылает события READ при отметке сообщения прочитанным через REST.
// Клиент отправляет новые сообщения в формате OrderMessageRequest, а получает сообщения типа models.OrderMessage и READ-события.
// @Tags orders
// @Param token query string true "access token"
// @Param id path string true "ID ордера"
// @Success 101 {object} models.OrderMessage "Switching Protocols"
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
		orderchat.AddClient(chat.ID, conn)
		defer func() {
			orderchat.RemoveClient(chat.ID, conn)
			conn.Close()
		}()

        if cache != nil {
            if history, err := cache.GetHistory(c.Request.Context(), chat.ID); err == nil {
                for _, m := range history {
                    if m.SenderName == "" {
                        var snd models.Client
                        if err := db.Select("username").Where("id = ?", m.ClientID).First(&snd).Error; err == nil {
                            m.SenderName = snd.Username
                        }
                    }
                    if err := orderchat.Send(conn, m); err != nil {
                        return
                    }
                }
            }
        }

		for {
			var r OrderMessageRequest
			if err := conn.ReadJSON(&r); err != nil {
				break
			}
			if r.Content == "" {
				continue
			}
        msg := models.OrderMessage{ChatID: chat.ID, ClientID: clientID, Type: models.MessageTypeText, Content: r.Content}
        if err := db.Create(&msg).Error; err != nil {
            continue
        }
        // подставляем имя отправителя
        var snd models.Client
        if err := db.Select("username").Where("id = ?", clientID).First(&snd).Error; err == nil {
            msg.SenderName = snd.Username
        }
        if cache != nil {
            _ = cache.AddMessage(c.Request.Context(), chat.ID, msg)
        }
			otherID := order.BuyerID
			if clientID == order.BuyerID {
				otherID = order.SellerID
			}
			if payload, err := json.Marshal(map[string]string{"orderId": order.ID, "messageId": msg.ID}); err == nil {
				n := models.Notification{ClientID: otherID, Type: "chat.message", Payload: payload, LinkTo: "/orders/" + order.ID}
				if err := db.Create(&n).Error; err == nil {
					notifications.Broadcast(otherID, n)
				}
			}
			orderchat.Broadcast(chat.ID, msg)
		}
	}
}
