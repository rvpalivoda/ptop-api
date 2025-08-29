package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// MarkPaidRequest тело запроса для отметки оплаты
type MarkPaidRequest struct {
	PaidAt *time.Time `json:"paidAt"`
}

// CancelOrderRequest тело запроса для отмены
type CancelOrderRequest struct {
	Reason *string `json:"reason"`
}

// DisputeRequest тело запроса для открытия спора
type DisputeRequest struct {
	Reason *string `json:"reason"`
}

// ResolveDisputeRequest тело запроса для решения спора
type ResolveDisputeRequest struct {
	Result  string  `json:"result"`
	Comment *string `json:"comment"`
}

// MarkOrderPaid godoc
// @Summary Отметить ордер оплаченным
// @Description WAIT_PAYMENT -> PAID. Только автор ордера. Отправляет уведомления и WS-событие.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID ордера"
// @Param input body handlers.MarkPaidRequest false "опционально: время оплаты"
// @Success 200 {object} models.OrderFull
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/paid [post]
func MarkOrderPaid(db *gorm.DB) gin.HandlerFunc {
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
		if order.Status != models.OrderStatusWaitPayment {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid status"})
			return
		}
		if order.AuthorID != clientID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		var r MarkPaidRequest
		_ = c.BindJSON(&r)
		when := time.Now()
		if r.PaidAt != nil {
			when = *r.PaidAt
		}
		res := db.Model(&models.Order{}).
			Where("id = ? AND status = ?", order.ID, models.OrderStatusWaitPayment).
			Updates(map[string]any{"status": models.OrderStatusPaid, "paid_at": when})
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if res.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status changed"})
			return
		}
		var full models.Order
		if err := db.Preload("Offer").
			Preload("Buyer").Preload("Seller").Preload("Author").Preload("OfferOwner").
			Preload("FromAsset").Preload("ToAsset").
			Preload("ClientPaymentMethod").
			Preload("ClientPaymentMethod.Country").
			Preload("ClientPaymentMethod.PaymentMethod").
			Where("id = ?", order.ID).First(&full).Error; err == nil {
			createOrderStatusNotifications(db, full)
			broadcastOrderStatus(full)
		}
		// возвращаем OrderFull
		var cpm *models.ClientPaymentMethod
		if full.ClientPaymentMethodID != "" {
			cpm = &full.ClientPaymentMethod
		}
		of := models.OrderFull{
			Order: full, Offer: full.Offer,
			Buyer: full.Buyer, Seller: full.Seller,
			Author: full.Author, OfferOwner: full.OfferOwner,
			FromAsset: full.FromAsset, ToAsset: full.ToAsset,
			ClientPaymentMethod: cpm,
		}
		c.JSON(http.StatusOK, of)
	}
}

// ReleaseOrder godoc
// @Summary Выпустить средства (завершить ордер)
// @Description PAID -> RELEASED. Только продавец (offerOwner). Устанавливает releasedAt, шлёт уведомления и WS.
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID ордера"
// @Success 200 {object} models.OrderFull
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/release [post]
func ReleaseOrder(db *gorm.DB) gin.HandlerFunc {
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
		if order.Status != models.OrderStatusPaid {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid status"})
			return
		}
		if order.OfferOwnerID != clientID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		now := time.Now()
		res := db.Model(&models.Order{}).
			Where("id = ? AND status = ?", order.ID, models.OrderStatusPaid).
			Updates(map[string]any{"status": models.OrderStatusReleased, "released_at": now})
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if res.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status changed"})
			return
		}
		var full models.Order
		if err := db.Preload("Offer").
			Preload("Buyer").Preload("Seller").Preload("Author").Preload("OfferOwner").
			Preload("FromAsset").Preload("ToAsset").
			Preload("ClientPaymentMethod").
			Preload("ClientPaymentMethod.Country").
			Preload("ClientPaymentMethod.PaymentMethod").
			Where("id = ?", order.ID).First(&full).Error; err == nil {
			createOrderStatusNotifications(db, full)
			broadcastOrderStatus(full)
		}
		var cpm *models.ClientPaymentMethod
		if full.ClientPaymentMethodID != "" {
			cpm = &full.ClientPaymentMethod
		}
		of := models.OrderFull{
			Order: full, Offer: full.Offer,
			Buyer: full.Buyer, Seller: full.Seller,
			Author: full.Author, OfferOwner: full.OfferOwner,
			FromAsset: full.FromAsset, ToAsset: full.ToAsset,
			ClientPaymentMethod: cpm,
		}
		c.JSON(http.StatusOK, of)
	}
}

// CancelOrder godoc
// @Summary Отменить ордер
// @Description WAIT_PAYMENT -> CANCELLED. Автор или продавец (при отсутствии оплаты). Шлёт уведомления и WS.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID ордера"
// @Param input body handlers.CancelOrderRequest false "причина отмены"
// @Success 200 {object} models.OrderFull
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/cancel [post]
func CancelOrder(db *gorm.DB) gin.HandlerFunc {
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
		if order.Status != models.OrderStatusWaitPayment {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid status"})
			return
		}
		if clientID != order.AuthorID && clientID != order.OfferOwnerID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		var r CancelOrderRequest
		_ = c.BindJSON(&r)
		upd := map[string]any{"status": models.OrderStatusCancelled}
		if r.Reason != nil {
			upd["cancel_reason"] = *r.Reason
		}
		res := db.Model(&models.Order{}).
			Where("id = ? AND status = ?", order.ID, models.OrderStatusWaitPayment).
			Updates(upd)
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if res.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status changed"})
			return
		}
		var full models.Order
		if err := db.Preload("Offer").
			Preload("Buyer").Preload("Seller").Preload("Author").Preload("OfferOwner").
			Preload("FromAsset").Preload("ToAsset").
			Preload("ClientPaymentMethod").
			Preload("ClientPaymentMethod.Country").
			Preload("ClientPaymentMethod.PaymentMethod").
			Where("id = ?", order.ID).First(&full).Error; err == nil {
			createOrderStatusNotifications(db, full)
			broadcastOrderStatus(full)
		}
		var cpm *models.ClientPaymentMethod
		if full.ClientPaymentMethodID != "" {
			cpm = &full.ClientPaymentMethod
		}
		of := models.OrderFull{
			Order: full, Offer: full.Offer,
			Buyer: full.Buyer, Seller: full.Seller,
			Author: full.Author, OfferOwner: full.OfferOwner,
			FromAsset: full.FromAsset, ToAsset: full.ToAsset,
			ClientPaymentMethod: cpm,
		}
		c.JSON(http.StatusOK, of)
	}
}

// OpenDispute godoc
// @Summary Открыть спор
// @Description PAID -> DISPUTE. Любая сторона. Шлёт уведомления и WS.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID ордера"
// @Param input body handlers.DisputeRequest false "причина"
// @Success 200 {object} models.OrderFull
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/dispute [post]
func OpenDispute(db *gorm.DB) gin.HandlerFunc {
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
		if order.Status != models.OrderStatusPaid {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid status"})
			return
		}
		if clientID != order.AuthorID && clientID != order.OfferOwnerID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		var r DisputeRequest
		_ = c.BindJSON(&r)
		upd := map[string]any{"status": models.OrderStatusDispute, "dispute_opened_at": time.Now()}
		if r.Reason != nil {
			upd["dispute_reason"] = *r.Reason
		}
		res := db.Model(&models.Order{}).
			Where("id = ? AND status = ?", order.ID, models.OrderStatusPaid).
			Updates(upd)
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if res.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status changed"})
			return
		}
		var full models.Order
		if err := db.Preload("Offer").
			Preload("Buyer").Preload("Seller").Preload("Author").Preload("OfferOwner").
			Preload("FromAsset").Preload("ToAsset").
			Preload("ClientPaymentMethod").
			Preload("ClientPaymentMethod.Country").
			Preload("ClientPaymentMethod.PaymentMethod").
			Where("id = ?", order.ID).First(&full).Error; err == nil {
			createOrderStatusNotifications(db, full)
			broadcastOrderStatus(full)
		}
		var cpm *models.ClientPaymentMethod
		if full.ClientPaymentMethodID != "" {
			cpm = &full.ClientPaymentMethod
		}
		of := models.OrderFull{
			Order: full, Offer: full.Offer,
			Buyer: full.Buyer, Seller: full.Seller,
			Author: full.Author, OfferOwner: full.OfferOwner,
			FromAsset: full.FromAsset, ToAsset: full.ToAsset,
			ClientPaymentMethod: cpm,
		}
		c.JSON(http.StatusOK, of)
	}
}

// ResolveDispute godoc
// @Summary Решить спор
// @Description DISPUTE -> RELEASED/CANCELLED. Только арбитраж. Шлёт уведомления и WS.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID ордера"
// @Param input body handlers.ResolveDisputeRequest true "результат спора"
// @Success 200 {object} models.OrderFull
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/dispute/resolve [post]
func ResolveDispute(db *gorm.DB) gin.HandlerFunc {
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
		if order.Status != models.OrderStatusDispute {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid status"})
			return
		}
		if clientID == order.AuthorID || clientID == order.OfferOwnerID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		var r ResolveDisputeRequest
		if err := c.BindJSON(&r); err != nil || (r.Result != string(models.OrderStatusReleased) && r.Result != string(models.OrderStatusCancelled)) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		upd := map[string]any{"status": models.OrderStatus(r.Result)}
		if r.Result == string(models.OrderStatusReleased) {
			upd["released_at"] = time.Now()
		} else {
			if r.Comment != nil {
				upd["cancel_reason"] = *r.Comment
			}
		}
		res := db.Model(&models.Order{}).
			Where("id = ? AND status = ?", order.ID, models.OrderStatusDispute).
			Updates(upd)
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if res.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status changed"})
			return
		}
		var full models.Order
		if err := db.Preload("Offer").
			Preload("Buyer").Preload("Seller").Preload("Author").Preload("OfferOwner").
			Preload("FromAsset").Preload("ToAsset").
			Preload("ClientPaymentMethod").
			Preload("ClientPaymentMethod.Country").
			Preload("ClientPaymentMethod.PaymentMethod").
			Where("id = ?", order.ID).First(&full).Error; err == nil {
			createOrderStatusNotifications(db, full)
			broadcastOrderStatus(full)
		}
		var cpm *models.ClientPaymentMethod
		if full.ClientPaymentMethodID != "" {
			cpm = &full.ClientPaymentMethod
		}
		of := models.OrderFull{
			Order: full, Offer: full.Offer,
			Buyer: full.Buyer, Seller: full.Seller,
			Author: full.Author, OfferOwner: full.OfferOwner,
			FromAsset: full.FromAsset, ToAsset: full.ToAsset,
			ClientPaymentMethod: cpm,
		}
		c.JSON(http.StatusOK, of)
	}
}
