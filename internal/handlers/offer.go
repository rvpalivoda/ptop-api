package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"ptop/internal/models"
)

type OfferRequest struct {
	MaxAmount              string   `json:"max_amount"`
	MinAmount              string   `json:"min_amount"`
	Amount                 string   `json:"amount"`
	Price                  string   `json:"price"`
	Type                   string   `json:"type"`
	FromAssetID            string   `json:"from_asset_id"`
	ToAssetID              string   `json:"to_asset_id"`
	Conditions             string   `json:"conditions"`
	OrderExpirationTimeout int      `json:"order_expiration_timeout"`
	ClientPaymentMethodIDs []string `json:"client_payment_method_ids"`
}

// CreateOffer godoc
// @Summary Создать объявление
// @Tags offers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body OfferRequest true "данные"
// @Success 200 {object} models.Offer
// @Failure 400 {object} ErrorResponse
// @Router /client/offers [post]
func CreateOffer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r OfferRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)

		if r.Type != models.OfferTypeBuy && r.Type != models.OfferTypeSell {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid type"})
			return
		}

		maxAmount, err := decimal.NewFromString(r.MaxAmount)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid max_amount"})
			return
		}
		minAmount, err := decimal.NewFromString(r.MinAmount)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid min_amount"})
			return
		}
		amount, err := decimal.NewFromString(r.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid amount"})
			return
		}
		price, err := decimal.NewFromString(r.Price)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid price"})
			return
		}
		timeout := r.OrderExpirationTimeout
		if timeout < 15 {
			timeout = 15
		}

		// load client payment methods
		pmSeen := map[string]struct{}{}
		var clientMethods []models.ClientPaymentMethod
		for _, id := range r.ClientPaymentMethodIDs {
			if _, exists := pmSeen[id]; exists {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "duplicate payment method"})
				return
			}
			pmSeen[id] = struct{}{}
			var m models.ClientPaymentMethod
			if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&m).Error; err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid client payment method"})
				return
			}
			clientMethods = append(clientMethods, m)
		}

		offer := models.Offer{
			MaxAmount:              maxAmount,
			MinAmount:              minAmount,
			Amount:                 amount,
			Price:                  price,
			Type:                   r.Type,
			FromAssetID:            r.FromAssetID,
			ToAssetID:              r.ToAssetID,
			Conditions:             r.Conditions,
			OrderExpirationTimeout: timeout,
			TTL:                    time.Now(),
			ClientID:               clientID,
		}
		if err := db.Create(&offer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if len(clientMethods) > 0 {
			if err := db.Model(&offer).Association("ClientPaymentMethods").Replace(clientMethods); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
				return
			}
		}
		c.JSON(http.StatusOK, offer)
	}
}

// UpdateOffer godoc
// @Summary Изменить объявление
// @Tags offers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param input body OfferRequest true "данные"
// @Success 200 {object} models.Offer
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /client/offers/{id} [put]
func UpdateOffer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var offer models.Offer
		if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&offer).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
			return
		}
		var r OfferRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		if r.Type != models.OfferTypeBuy && r.Type != models.OfferTypeSell {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid type"})
			return
		}
		maxAmount, err := decimal.NewFromString(r.MaxAmount)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid max_amount"})
			return
		}
		minAmount, err := decimal.NewFromString(r.MinAmount)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid min_amount"})
			return
		}
		amount, err := decimal.NewFromString(r.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid amount"})
			return
		}
		price, err := decimal.NewFromString(r.Price)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid price"})
			return
		}
		timeout := r.OrderExpirationTimeout
		if timeout < 15 {
			timeout = 15
		}

		pmSeen := map[string]struct{}{}
		var clientMethods []models.ClientPaymentMethod
		for _, pid := range r.ClientPaymentMethodIDs {
			if _, ok := pmSeen[pid]; ok {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "duplicate payment method"})
				return
			}
			pmSeen[pid] = struct{}{}
			var m models.ClientPaymentMethod
			if err := db.Where("id = ? AND client_id = ?", pid, clientID).First(&m).Error; err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid client payment method"})
				return
			}
			clientMethods = append(clientMethods, m)
		}
		offer.MaxAmount = maxAmount
		offer.MinAmount = minAmount
		offer.Amount = amount
		offer.Price = price
		offer.Type = r.Type
		offer.FromAssetID = r.FromAssetID
		offer.ToAssetID = r.ToAssetID
		offer.Conditions = r.Conditions
		offer.OrderExpirationTimeout = timeout
		if err := db.Save(&offer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if err := db.Model(&offer).Association("ClientPaymentMethods").Replace(clientMethods); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, offer)
	}
}

// EnableOffer godoc
// @Summary Включить объявление
// @Tags offers
// @Security BearerAuth
// @Param id path string true "ID"
// @Success 200 {object} models.Offer
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /client/offers/{id}/enable [post]
func EnableOffer(db *gorm.DB, maxActive int) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var offer models.Offer
		if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&offer).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
			return
		}
		if offer.IsEnabled {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "already enabled"})
			return
		}
		var count int64
		db.Model(&models.Offer{}).Where("client_id = ? AND is_enabled = ? AND ttl > ?", clientID, true, time.Now()).Count(&count)
		if count >= int64(maxActive) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "max active offers reached"})
			return
		}
		now := time.Now()
		offer.IsEnabled = true
		offer.EnabledAt = &now
		offer.DisabledAt = nil
		offer.TTL = now.Add(30 * 24 * time.Hour)
		if err := db.Save(&offer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, offer)
	}
}

// DisableOffer godoc
// @Summary Отключить объявление
// @Tags offers
// @Security BearerAuth
// @Param id path string true "ID"
// @Success 200 {object} models.Offer
// @Failure 404 {object} ErrorResponse
// @Router /client/offers/{id}/disable [post]
func DisableOffer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var offer models.Offer
		if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&offer).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
			return
		}
		now := time.Now()
		offer.IsEnabled = false
		offer.DisabledAt = &now
		offer.EnabledAt = nil
		offer.TTL = now
		if err := db.Save(&offer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, offer)
	}
}

// ListOffers godoc
// @Summary Список активных объявлений
// @Tags offers
// @Security BearerAuth
// @Produce json
// @Param from_asset query string false "ID актива от"
// @Param to_asset query string false "ID актива к"
// @Param min_amount query string false "минимальная сумма"
// @Param max_amount query string false "максимальная сумма"
// @Param payment_method query string false "ID способа оплаты"
// @Param type query string false "тип объявления: buy или sell"
// @Success 200 {array} models.Offer
// @Router /offers [get]
func ListOffers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&models.Offer{}).Where("is_enabled = ? AND ttl > ?", true, time.Now()).Distinct()
		if fa := c.Query("from_asset"); fa != "" {
			query = query.Where("from_asset_id = ?", fa)
		}
		if ta := c.Query("to_asset"); ta != "" {
			query = query.Where("to_asset_id = ?", ta)
		}
		if min := c.Query("min_amount"); min != "" {
			v, err := decimal.NewFromString(min)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid min_amount"})
				return
			}
			query = query.Where("max_amount >= ?", v)
		}
		if max := c.Query("max_amount"); max != "" {
			v, err := decimal.NewFromString(max)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid max_amount"})
				return
			}
			query = query.Where("min_amount <= ?", v)
		}
		if pm := c.Query("payment_method"); pm != "" {
			query = query.Joins("JOIN offer_client_payment_methods ocpm ON ocpm.offer_id = offers.id").
				Joins("JOIN client_payment_methods cpm ON cpm.id = ocpm.client_payment_method_id").
				Where("cpm.payment_method_id = ?", pm)
		}
		if t := c.Query("type"); t != "" {
			if t != models.OfferTypeBuy && t != models.OfferTypeSell {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid type"})
				return
			}
			query = query.Where("type = ?", t)
		}
		var offers []models.Offer
		if err := query.Find(&offers).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, offers)
	}
}

// ListClientOffers godoc
// @Summary Список объявлений клиента
// @Tags offers
// @Security BearerAuth
// @Produce json
// @Param enabled query bool false "фильтр по активным"
// @Success 200 {array} models.Offer
// @Router /client/offers [get]
func ListClientOffers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		query := db.Where("client_id = ?", clientID)
		if en := c.Query("enabled"); en != "" {
			if en == "true" {
				query = query.Where("is_enabled = ?", true)
			} else if en == "false" {
				query = query.Where("is_enabled = ?", false)
			}
		}
		var offers []models.Offer
		if err := query.Find(&offers).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, offers)
	}
}
