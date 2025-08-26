package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/notifications"
)

// NotificationsWS godoc
// @Summary Websocket уведомлений
// @Description Подключает клиента к потоку уведомлений. После подключения сервер отправляет непрочитанные уведомления.
// @Tags notifications
// @Param token query string true "access token"
// @Success 101 {object} models.Notification "Switching Protocols"
// @Failure 401 {object} ErrorResponse
// @Router /ws/notifications [get]
func NotificationsWS(db *gorm.DB) gin.HandlerFunc {
	notifications.SetDB(db)
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		notifications.AddClient(clientID, conn)
		defer func() {
			notifications.RemoveClient(clientID, conn)
			conn.Close()
		}()

		var list []models.Notification
		if err := db.Where("client_id = ? AND read_at IS NULL AND sent_at IS NULL", clientID).Find(&list).Error; err == nil {
			for _, n := range list {
				if err := notifications.Send(conn, n); err != nil {
					return
				}
			}
		}

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}
