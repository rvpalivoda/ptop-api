package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// ListClientTransactionsIn godoc
// @Summary Список входящих транзакций клиента
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Param limit query int false "лимит"
// @Param offset query int false "смещение"
// @Success 200 {array} models.TransactionIn
// @Router /client/transactions/in [get]
func ListClientTransactionsIn(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		limit, offset := parsePagination(c)
		var txs []models.TransactionIn
		if err := db.Model(&models.TransactionIn{}).
			Select("transaction_ins.*, assets.name as asset_name").
			Joins("LEFT JOIN assets ON assets.id = transaction_ins.asset_id").
			Where("transaction_ins.client_id = ?", clientID).
			Order("transaction_ins.created_at desc").
			Limit(limit).Offset(offset).Find(&txs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, txs)
	}
}

// ListClientTransactionsOut godoc
// @Summary Список исходящих транзакций клиента
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Param limit query int false "лимит"
// @Param offset query int false "смещение"
// @Success 200 {array} models.TransactionOut
// @Router /client/transactions/out [get]
func ListClientTransactionsOut(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		limit, offset := parsePagination(c)
		var txs []models.TransactionOut
		if err := db.Model(&models.TransactionOut{}).
			Select("transaction_outs.*, assets.name as asset_name").
			Joins("LEFT JOIN assets ON assets.id = transaction_outs.asset_id").
			Where("transaction_outs.client_id = ?", clientID).
			Order("transaction_outs.created_at desc").
			Limit(limit).Offset(offset).Find(&txs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, txs)
	}
}

// ListClientTransactionsInternal godoc
// @Summary Список внутренних транзакций клиента
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Param limit query int false "лимит"
// @Param offset query int false "смещение"
// @Success 200 {array} models.TransactionInternal
// @Router /client/transactions/internal [get]
func ListClientTransactionsInternal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		limit, offset := parsePagination(c)
		var txs []models.TransactionInternal
		if err := db.Model(&models.TransactionInternal{}).
			Select("transaction_internals.*, assets.name as asset_name").
			Joins("LEFT JOIN assets ON assets.id = transaction_internals.asset_id").
			Where("transaction_internals.from_client_id = ? OR transaction_internals.to_client_id = ?", clientID, clientID).
			Order("transaction_internals.created_at desc").
			Limit(limit).Offset(offset).Find(&txs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, txs)
	}
}
