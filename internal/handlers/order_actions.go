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
    // Возможные действия: markPaid, cancel, dispute, release
    // Будет сериализован как массив строк
    Actions []models.OrderAction `json:"actions" swaggertype:"array,string" enums:"markPaid,cancel,dispute,release"`
}

// GetOrderActions godoc
// @Summary Доступные действия по ордеру
// @Description Возвращает список действий, которые может выполнить текущий пользователь над ордером в зависимости от его роли и текущего статуса.
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID ордера"
// @Success 200 {object} OrderActionsResponse
    // @Failure 401 {object} ErrorResponse
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

        actions := []models.OrderAction{}
        switch {
        case order.AuthorID == clientID:
            // buyer actions
            switch order.Status {
            case models.OrderStatusWaitPayment:
                actions = append(actions, models.OrderActionMarkPaid, models.OrderActionCancel)
            case models.OrderStatusPaid:
                actions = append(actions, models.OrderActionDispute)
            }
        case order.OfferOwnerID == clientID:
            // seller actions
            switch order.Status {
            case models.OrderStatusWaitPayment:
                actions = append(actions, models.OrderActionCancel)
            case models.OrderStatusPaid:
                actions = append(actions, models.OrderActionRelease, models.OrderActionDispute)
            }
        default:
            c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
            return
        }

		c.JSON(http.StatusOK, OrderActionsResponse{Actions: actions})
	}
}
