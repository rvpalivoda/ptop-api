package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"ptop/internal/models"
)

type OrderRequest struct {
	OfferID         string `json:"offer_id"`
	Amount          string `json:"amount"`
	PaymentMethodID string `json:"payment_method_id"`
}

// CreateOrder godoc
// @Summary Создать ордер
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body OrderRequest true "данные"
// @Success 200 {object} models.Order
// @Failure 400 {object} ErrorResponse
// @Router /client/orders [post]
func CreateOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r OrderRequest
		if err := c.BindJSON(&r); err != nil || r.OfferID == "" || r.Amount == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		amt, err := decimal.NewFromString(r.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid amount"})
			return
		}
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var offer models.Offer
		if err := db.Preload("FromAsset").Preload("ToAsset").Where("id = ?", r.OfferID).First(&offer).Error; err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid offer"})
			return
		}
		order := models.Order{
			OfferID:         offer.ID,
			BuyerID:         clientID,
			SellerID:        offer.ClientID,
			FromAssetID:     offer.FromAssetID,
			ToAssetID:       offer.ToAssetID,
			Amount:          amt,
			Price:           offer.Price,
			PaymentMethodID: r.PaymentMethodID,
			Status:          models.OrderStatusWaitPayment,
			ExpiresAt:       time.Now().Add(time.Duration(offer.OrderExpirationTimeout) * time.Minute),
		}
		if offer.FromAsset.Type == "crypto" || offer.ToAsset.Type == "crypto" {
			order.IsEscrow = true
		}
		if err := db.Create(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, order)
	}
}

// ListClientOrders godoc
// @Summary Список ордеров клиента
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Order
// @Router /client/orders [get]
func ListClientOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var orders []models.Order
		if err := db.Where("buyer_id = ?", clientID).Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}
