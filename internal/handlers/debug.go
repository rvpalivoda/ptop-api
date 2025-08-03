package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"ptop/internal/models"
)

type DebugDepositor interface {
	TriggerDeposit(walletID string, amount decimal.Decimal)
}

type DebugDepositRequest struct {
	WalletID string `json:"wallet_id" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
}

// DebugDeposit godoc
// @Summary      Тестовый депозит
// @Description  Создаёт фейковый депозит на указанный кошелёк
// @Tags         debug
// @Accept       json
// @Produce      json
// @Param        request body DebugDepositRequest true "Запрос"
// @Success      204
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /debug/deposit [post]
func DebugDeposit(db *gorm.DB, watchers map[string]DebugDepositor) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DebugDepositRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		amt, err := decimal.NewFromString(req.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
			return
		}
		var wal models.Wallet
		if err := db.Preload("Asset").Where("id = ?", req.WalletID).First(&wal).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		var watcher DebugDepositor
		name := strings.ToUpper(wal.Asset.Name)
		switch {
		case strings.HasPrefix(name, "BTC"):
			watcher = watchers["BTC"]
		case strings.HasPrefix(name, "ETH"):
			watcher = watchers["ETH"]
		case strings.HasPrefix(name, "XMR"):
			watcher = watchers["XMR"]
		}
		if watcher == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "watcher not available"})
			return
		}
		watcher.TriggerDeposit(wal.ID, amt)
		c.Status(http.StatusNoContent)
	}
}
