package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// OrderMessageRequest используется для текстовых сообщений.
type OrderMessageRequest struct {
	Content string `json:"content"`
}

// ListOrderMessages godoc
// @Summary Список сообщений ордера
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID ордера"
// @Param cursor query string false "cursor"
// @Param after query string false "after"
// @Success 200 {array} models.OrderMessage
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/messages [get]
func ListOrderMessages(db *gorm.DB) gin.HandlerFunc {
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
				c.JSON(http.StatusOK, []models.OrderMessage{})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		q := db.Where("chat_id = ?", chat.ID)
		if cursor := c.Query("cursor"); cursor != "" {
			q = q.Where("id > ?", cursor)
		}
		if after := c.Query("after"); after != "" {
			if t, err := time.Parse(time.RFC3339, after); err == nil {
				q = q.Where("created_at > ?", t)
			}
		}
		var msgs []models.OrderMessage
		if err := q.Order("created_at asc").Limit(50).Find(&msgs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, msgs)
	}
}

// CreateOrderMessage godoc
// @Summary Отправить сообщение в ордер
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "ID ордера"
// @Param input body OrderMessageRequest false "данные"
// @Param file formData file false "файл"
// @Success 200 {object} models.OrderMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/messages [post]
func CreateOrderMessage(db *gorm.DB) gin.HandlerFunc {
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
		var msg models.OrderMessage
		if c.ContentType() == "multipart/form-data" {
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid file"})
				return
			}
			msg = models.OrderMessage{ChatID: chat.ID, ClientID: clientID, Type: models.MessageTypeFile, Content: file.Filename}
		} else {
			var r OrderMessageRequest
			if err := c.BindJSON(&r); err != nil || r.Content == "" {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
				return
			}
			msg = models.OrderMessage{ChatID: chat.ID, ClientID: clientID, Type: models.MessageTypeText, Content: r.Content}
		}
		if err := db.Create(&msg).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, msg)
	}
}

// ReadOrderMessage godoc
// @Summary Отметить сообщение прочитанным
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID ордера"
// @Param msgId path string true "ID сообщения"
// @Success 200 {object} models.OrderMessage
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/messages/{msgId}/read [patch]
func ReadOrderMessage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		msgID := c.Param("msgId")
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
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid message"})
			return
		}
		var msg models.OrderMessage
		if err := db.Where("id = ? AND chat_id = ?", msgID, chat.ID).First(&msg).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid message"})
			return
		}
		now := time.Now()
		if err := db.Model(&msg).Update("read_at", now).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		msg.ReadAt = &now
		c.JSON(http.StatusOK, msg)
	}
}
