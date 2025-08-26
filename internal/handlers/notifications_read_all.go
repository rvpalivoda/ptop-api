package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// NotificationsReadAllResponse ответ на массовое чтение уведомлений.
type NotificationsReadAllResponse struct {
	Count int `json:"count"`
}

// ReadAllNotifications godoc
// @Summary Отметить все уведомления прочитанными
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} handlers.NotificationsReadAllResponse
// @Router /notifications/read-all [post]
func ReadAllNotifications(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		now := time.Now()
		res := db.Model(&models.Notification{}).
			Where("client_id = ? AND read_at IS NULL", clientID).
			Update("read_at", now)
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, NotificationsReadAllResponse{Count: int(res.RowsAffected)})
	}
}
