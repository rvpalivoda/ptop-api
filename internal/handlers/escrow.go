package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// ListClientEscrows godoc
// @Summary Список эскроу клиента
// @Tags escrows
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Escrow
// @Router /client/escrows [get]
func ListClientEscrows(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var escrows []models.Escrow
		if err := db.Where("client_id = ?", clientID).Find(&escrows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, escrows)
	}
}
