package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// GetCountries godoc
// @Summary Список стран
// @Tags reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Country
// @Router /countries [get]
func GetCountries(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var countries []models.Country
		if err := db.Find(&countries).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, countries)
	}
}

// GetPaymentMethods godoc
// @Summary Список платёжных методов
// @Tags reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.PaymentMethod "Расширенная структура платёжных методов"
// @Router /payment-methods [get]
func GetPaymentMethods(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var methods []models.PaymentMethod
		if err := db.Find(&methods).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, methods)
	}
}

// AssetWithWallet включает актив и адрес кошелька клиента.
type AssetWithWallet struct {
	models.Asset
	Value string `json:"value"`
}

// GetAssets godoc
// @Summary Список активных активов
// @Tags reference
// @Produce json
// @Success 200 {array} models.Asset
// @Router /assets [get]
func GetAssets(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var assets []models.Asset
		if err := db.Where("is_active = ?", true).Find(&assets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, assets)
	}
}

// GetClientAssets godoc
// @Summary Список активных активов с адресами кошельков клиента
// @Tags reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} handlers.AssetWithWallet
// @Router /client/assets [get]
func GetClientAssets(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var assets []AssetWithWallet
		if err := db.Model(&models.Asset{}).
			Select("assets.id, assets.name, assets.type, assets.is_active, assets.is_convertible, COALESCE(wallets.value, '') AS value").
			Joins("LEFT JOIN wallets ON wallets.asset_id = assets.id AND wallets.client_id = ? AND wallets.is_enabled = ?", clientID, true).
			Where("assets.is_active = ?", true).
			Scan(&assets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, assets)
	}
}
