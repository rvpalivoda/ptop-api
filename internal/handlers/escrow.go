package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"ptop/internal/models"
)

type EscrowDetail struct {
	ID        string          `json:"id"`
	Client    models.Client   `json:"client"`
	Asset     models.Asset    `json:"asset"`
	Amount    decimal.Decimal `json:"amount"`
	Offer     *models.Offer   `json:"offer,omitempty"`
	Order     *models.Order   `json:"order,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type EscrowNavResponse struct {
	Escrow EscrowDetail `json:"escrow"`
	NextID string       `json:"nextId,omitempty"`
	PrevID string       `json:"prevId,omitempty"`
}

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

// GetClientEscrow godoc
// @Summary Просмотр эскроу клиента
// @Tags escrows
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID эскроу"
// @Param dir query string false "навигация: next или prev"
// @Success 200 {object} EscrowNavResponse
// @Failure 404 {object} ErrorResponse
// @Router /client/escrows/{id} [get]
func GetClientEscrow(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		id := c.Param("id")
		dir := c.Query("dir")

		var esc models.Escrow
		q := db.Preload("Client").Preload("Asset").Preload("Offer").Preload("Order").Where("client_id = ?", clientID)
		switch dir {
		case "next":
			var base models.Escrow
			if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&base).Error; err != nil {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
				return
			}
			if err := q.Where("created_at > ?", base.CreatedAt).Order("created_at asc").First(&esc).Error; err != nil {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
				return
			}
		case "prev":
			var base models.Escrow
			if err := db.Where("id = ? AND client_id = ?", id, clientID).First(&base).Error; err != nil {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
				return
			}
			if err := q.Where("created_at < ?", base.CreatedAt).Order("created_at desc").First(&esc).Error; err != nil {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
				return
			}
		default:
			if err := q.Where("id = ?", id).First(&esc).Error; err != nil {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
				return
			}
		}

		var nextEsc, prevEsc models.Escrow
		nextID, prevID := "", ""
		if err := db.Where("client_id = ? AND created_at > ?", clientID, esc.CreatedAt).Order("created_at asc").Select("id").First(&nextEsc).Error; err == nil {
			nextID = nextEsc.ID
		}
		if err := db.Where("client_id = ? AND created_at < ?", clientID, esc.CreatedAt).Order("created_at desc").Select("id").First(&prevEsc).Error; err == nil {
			prevID = prevEsc.ID
		}

		resp := EscrowNavResponse{
			Escrow: EscrowDetail{
				ID:        esc.ID,
				Client:    esc.Client,
				Asset:     esc.Asset,
				Amount:    esc.Amount,
				CreatedAt: esc.CreatedAt,
				UpdatedAt: esc.UpdatedAt,
			},
			NextID: nextID,
			PrevID: prevID,
		}
		if esc.OfferID != nil {
			resp.Escrow.Offer = &esc.Offer
		}
		if esc.OrderID != nil {
			resp.Escrow.Order = &esc.Order
		}

		c.JSON(http.StatusOK, resp)
	}
}
