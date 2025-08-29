package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// OrderActionsResponse ответ со списком доступных действий
// @Description Список действий, доступных текущему пользователю по данному ордеру.
type OrderActionsResponse struct {
	Actions []string `json:"actions"`
}

// GetOrderActions godoc
// @Summary Доступные действия по ордеру
// @Description Возвращает список действий, которые может выполнить текущий пользователь над ордером в зависимости от его роли и текущего статуса.
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID ордера"
// @Success 200 {object} OrderActionsResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/actions [get]
func GetOrderActions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)

		var order models.Order
		if err := db.Select("id", "status", "author_id", "offer_owner_id").Where("id = ?", orderID).First(&order).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid order"})
			return
		}

		actions := []string{}
		switch {
		case order.AuthorID == clientID:
			// buyer actions
			switch order.Status {
			case models.OrderStatusWaitPayment:
				actions = append(actions, "markPaid", "cancel")
			case models.OrderStatusPaid:
				actions = append(actions, "dispute")
			}
		case order.OfferOwnerID == clientID:
			// seller actions
			switch order.Status {
			case models.OrderStatusWaitPayment:
				actions = append(actions, "cancel")
			case models.OrderStatusPaid:
				actions = append(actions, "release", "dispute")
			}
		default:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}

		c.JSON(http.StatusOK, OrderActionsResponse{Actions: actions})
	}
}
