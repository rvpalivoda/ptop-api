package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
)

func parsePagination(c *gin.Context) (limit, offset int) {
	limit = 50
	offset = 0
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if oStr := c.Query("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}
	return
}

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
		if err := db.Where("client_id = ?", clientID).Order("created_at desc").Limit(limit).Offset(offset).Find(&txs).Error; err != nil {
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
		if err := db.Where("client_id = ?", clientID).Order("created_at desc").Limit(limit).Offset(offset).Find(&txs).Error; err != nil {
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
		if err := db.Where("from_client_id = ? OR to_client_id = ?", clientID, clientID).Order("created_at desc").Limit(limit).Offset(offset).Find(&txs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, txs)
	}
}
