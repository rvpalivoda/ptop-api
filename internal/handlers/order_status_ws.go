package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/notifications"
)

// OrderStatusEvent уведомление об изменении статуса ордера.
// Type всегда `order.status_changed`.
type OrderStatusEvent struct {
	Type  string           `json:"type" example:"order.status_changed"`
	Order models.OrderFull `json:"order"`
}

var orderStatusClients = struct {
	sync.RWMutex
	m map[string]map[*websocket.Conn]bool
}{m: make(map[string]map[*websocket.Conn]bool)}

func sendOrderStatusEvent(conn *websocket.Conn, ord models.OrderFull) error {
	return conn.WriteJSON(OrderStatusEvent{Type: "order.status_changed", Order: ord})
}

func createOrderStatusNotifications(db *gorm.DB, ord models.Order) {
	payload, err := json.Marshal(map[string]string{
		"orderId": ord.ID,
		"status":  string(ord.Status),
	})
	if err != nil {
		return
	}
	for _, cid := range []string{ord.AuthorID, ord.OfferOwnerID} {
		n := models.Notification{ClientID: cid, Type: "order.status_changed", Payload: payload}
		if err := db.Create(&n).Error; err == nil {
			notifications.Broadcast(cid, n)
		}
	}
}

func broadcastOrderStatus(order models.Order) {
	ofull := models.OrderFull{
		Order:      order,
		Offer:      order.Offer,
		Buyer:      order.Buyer,
		Seller:     order.Seller,
		Author:     order.Author,
		OfferOwner: order.OfferOwner,
		FromAsset:  order.FromAsset,
		ToAsset:    order.ToAsset,
	}
	if order.ClientPaymentMethodID != "" {
		ofull.ClientPaymentMethod = &order.ClientPaymentMethod
	}
	orderStatusClients.Lock()
	conns := orderStatusClients.m[order.ID]
	for c := range conns {
		if err := sendOrderStatusEvent(c, ofull); err != nil {
			c.Close()
			delete(conns, c)
		}
	}
	orderStatusClients.Unlock()
}

// OrderStatusWS godoc
// @Summary Websocket уведомлений о статусе ордера
// @Description Позволяет автору и владельцу оффера получать события OrderStatusEvent при каждом изменении статуса указанного ордера.
// @Tags orders
// @Security BearerAuth
// @Param id path string true "ID ордера"
// @Success 101 {object} handlers.OrderStatusEvent "Switching Protocols"
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /ws/orders/{id}/status [get]
func OrderStatusWS(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var order models.Order
		if err := db.Preload("Offer").
			Preload("Buyer").
			Preload("Seller").
			Preload("Author").
			Preload("OfferOwner").
			Preload("FromAsset").
			Preload("ToAsset").
			Preload("ClientPaymentMethod").
			Preload("ClientPaymentMethod.Country").
			Preload("ClientPaymentMethod.PaymentMethod").
			Where("id = ?", orderID).First(&order).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid order"})
			return
		}
		if clientID != order.AuthorID && clientID != order.OfferOwnerID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		orderStatusClients.Lock()
		conns, ok := orderStatusClients.m[order.ID]
		if !ok {
			conns = make(map[*websocket.Conn]bool)
			orderStatusClients.m[order.ID] = conns
		}
		conns[conn] = true
		orderStatusClients.Unlock()
		defer func() {
			orderStatusClients.Lock()
			delete(orderStatusClients.m[order.ID], conn)
			orderStatusClients.Unlock()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}
