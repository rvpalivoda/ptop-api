package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/orderchat"
	"ptop/internal/services"
)

// OrderChatWS godoc
// @Summary Websocket чат ордера
// @Description Подключает покупателя и продавца к чату ордера.
// После подключения сервер отправляет историю сообщений (models.OrderMessage).
// Клиент отправляет новые сообщения в формате OrderMessageRequest, а получает сообщения типа models.OrderMessage.
// @Tags orders
// @Security BearerAuth
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
			if cache != nil {
				_ = cache.AddMessage(c.Request.Context(), chat.ID, msg)
			}
			orderchat.Broadcast(chat.ID, msg)
		}
	}
}
