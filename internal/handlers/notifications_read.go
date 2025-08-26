package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// ReadNotification godoc
// @Summary Отметить уведомление прочитанным
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID уведомления"
// @Success 200 {object} models.Notification
// @Failure 404 {object} ErrorResponse
// @Router /notifications/{id}/read [patch]
func ReadNotification(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		id := c.Param("id")
		var n models.Notification
		if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&n).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid notification"})
			return
		}
		now := time.Now()
		if err := db.Model(&n).Update("read_at", now).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		n.ReadAt = &now
		c.JSON(http.StatusOK, n)
	}
}
