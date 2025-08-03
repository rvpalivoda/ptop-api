package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/services"
)

type WalletRequest struct {
	AssetID string `json:"asset_id"`
}

// CreateWallet godoc
// @Summary Создать кошелёк
// @Description При DEBUG_FAKE_NETWORK=true адрес генерируется без обращения к сети в формате fake:{asset}:{client}:{index}
// @Tags wallets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body WalletRequest true "данные"
// @Success 200 {object} models.Wallet
// @Failure 400 {object} ErrorResponse
// @Router /client/wallets [post]
func CreateWallet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r WalletRequest
		if err := c.BindJSON(&r); err != nil || r.AssetID == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var asset models.Asset
		if err := db.Where("id = ? AND type = ?", r.AssetID, models.AssetTypeCrypto).First(&asset).Error; err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid asset"})
			return
		}
		var count int64
		db.Model(&models.Wallet{}).Where("client_id = ? AND asset_id = ? AND is_enabled = ?", clientID, r.AssetID, true).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "wallet exists"})
			return
		}
		// в режиме DEBUG_FAKE_NETWORK возвращает фейковый адрес
		val, idx, err := services.GetAddress(db, clientID, r.AssetID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "address error"})
			return
		}
		now := time.Now()
		w := models.Wallet{
			ClientID:        clientID,
			AssetID:         r.AssetID,
			Value:           val,
			DerivationIndex: idx,
			IsEnabled:       true,
			EnabledAt:       now,
		}
		if err := db.Create(&w).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		b := models.Balance{ClientID: clientID, AssetID: r.AssetID, Amount: decimal.Zero, AmountEscrow: decimal.Zero}
		if err := db.Create(&b).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, w)
	}
}

// ListClientWallets godoc
// @Summary Список кошельков клиента
// @Tags wallets
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Wallet
// @Router /client/wallets [get]
func ListClientWallets(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var wallets []models.Wallet
		if err := db.Where("client_id = ? AND is_enabled = ?", clientID, true).Find(&wallets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, wallets)
	}
}
