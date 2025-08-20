package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

type CreateClientPaymentMethodRequest struct {
	CountryID           string `json:"country_id"`
	PaymentMethodID     string `json:"payment_method_id"`
	City                string `json:"city"`
	PostCode            string `json:"post_code"`
	DetailedInformation string `json:"detailed_information"`
	Name                string `json:"name"`
}

// ListClientPaymentMethods godoc
// @Summary Список платёжных методов клиента
// @Tags client-payment-methods
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.ClientPaymentMethod
// @Router /client/payment-methods [get]
func ListClientPaymentMethods(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var methods []models.ClientPaymentMethod
		if err := db.Where("client_id = ?", clientID).Preload("Country").Preload("PaymentMethod").Find(&methods).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, methods)
	}
}

// CreateClientPaymentMethod godoc
// @Summary Создать платёжный метод клиента
// @Tags client-payment-methods
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body CreateClientPaymentMethodRequest true "данные"
// @Success 200 {object} models.ClientPaymentMethod
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /client/payment-methods [post]
func CreateClientPaymentMethod(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r CreateClientPaymentMethodRequest
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
		var count int64
		db.Model(&models.ClientPaymentMethod{}).Where("client_id = ? AND name = ?", clientID, r.Name).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "name exists"})
			return
		}
		m := models.ClientPaymentMethod{
			ClientID:            clientID,
			CountryID:           r.CountryID,
			PaymentMethodID:     r.PaymentMethodID,
			City:                r.City,
			PostCode:            r.PostCode,
			DetailedInformation: r.DetailedInformation,
			Name:                r.Name,
		}
		if err := db.Create(&m).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if err := db.Preload("Country").Preload("PaymentMethod").First(&m, "id = ?", m.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, m)
	}
}

// UpdateClientPaymentMethod godoc
// @Summary Изменить платёжный метод клиента
// @Tags client-payment-methods
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param input body CreateClientPaymentMethodRequest true "данные"
// @Success 200 {object} models.ClientPaymentMethod
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /client/payment-methods/{id} [put]
func UpdateClientPaymentMethod(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var m models.ClientPaymentMethod
		if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&m).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
			return
		}
		var r CreateClientPaymentMethodRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		var count int64
		db.Model(&models.ClientPaymentMethod{}).Where("client_id = ? AND name = ? AND id <> ?", clientID, r.Name, m.ID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "name exists"})
			return
		}
		m.CountryID = r.CountryID
		m.PaymentMethodID = r.PaymentMethodID
		m.City = r.City
		m.PostCode = r.PostCode
		m.DetailedInformation = r.DetailedInformation
		m.Name = r.Name
		if err := db.Save(&m).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		if err := db.Preload("Country").Preload("PaymentMethod").First(&m, "id = ?", m.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, m)
	}
}

// DeleteClientPaymentMethod godoc
// @Summary Удалить платёжный метод клиента
// @Tags client-payment-methods
// @Security BearerAuth
// @Param id path string true "ID"
// @Success 200 {object} StatusResponse
// @Failure 404 {object} ErrorResponse
// @Router /client/payment-methods/{id} [delete]
func DeleteClientPaymentMethod(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		id := c.Param("id")
		var m models.ClientPaymentMethod
		if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&m).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
			return
		}
		if err := db.Delete(&m).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, StatusResponse{Status: "deleted"})
	}
}
