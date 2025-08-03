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
// @Success 200 {array} models.PaymentMethod
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

// GetAssets godoc
// @Summary Список активных активов
// @Tags reference
// @Security BearerAuth
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
