package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// ListClientBalances godoc
// @Summary Список балансов клиента
// @Tags balances
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Balance
// @Router /client/balances [get]
func ListClientBalances(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var balances []models.Balance
		if err := db.Where("client_id = ?", clientID).Find(&balances).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, balances)
	}
}
