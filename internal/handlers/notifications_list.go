package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// ListNotifications godoc
// @Summary Список уведомлений клиента
// @Tags notifications
// @Security BearerAuth
// @Produce json
// @Param limit query int false "лимит"
// @Param offset query int false "смещение"
// @Success 200 {array} models.Notification
// @Router /notifications [get]
func ListNotifications(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		limit, offset := parsePagination(c)
		var ns []models.Notification
		if err := db.Where("client_id = ?", clientID).
			Order("created_at desc").
			Limit(limit).Offset(offset).Find(&ns).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, ns)
	}
}
