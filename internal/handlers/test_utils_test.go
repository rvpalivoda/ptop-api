package handlers

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/services"
	storage "ptop/internal/services/storage"
)

// setupTest создаёт in-memory БД и маршруты для тестов.
type dummyStorage struct{}

func (d *dummyStorage) Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) (string, error) {
	return objectName, nil
}

func (d *dummyStorage) GetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	return "https://example.com/" + objectName, nil
}

var _ storage.Storage = (*dummyStorage)(nil)

func setupTest(t *testing.T) (*gorm.DB, *gin.Engine, map[string]time.Duration) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Client{},
		&models.Token{},
		&models.Country{},
		&models.PaymentMethod{},
		&models.ClientPaymentMethod{},
		&models.Asset{},
		&models.Offer{},
		&models.Wallet{},
		&models.Balance{},
		&models.Escrow{},
		&models.Order{},
		&models.OrderChat{},
		&models.OrderMessage{},
		&models.TransactionIn{},
		&models.TransactionOut{},
		&models.TransactionInternal{},
		&models.Notification{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ttl := map[string]time.Duration{"access": time.Minute, "refresh": time.Hour}

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cache := services.NewChatCache(rdb, 50)
	store := &dummyStorage{}

	r := gin.Default()
	auth := r.Group("/auth")
	auth.POST("/register", Register(db, ttl))
	auth.POST("/login", Login(db, ttl))
	auth.POST("/refresh", Refresh(db, ttl))
	auth.GET("/recover/:username", RecoverChallenge(db))
	auth.POST("/recover", Recover(db, ttl))
	auth.Use(AuthMiddleware(db))
	auth.POST("/logout", Logout(db))
	auth.GET("/profile", Profile(db))
	auth.POST("/username", ChangeUsername(db))
	auth.POST("/pincode", SetPinCode(db))
	auth.POST("/2fa/enable", Enable2FA(db))
	auth.POST("/2fa/disable", Disable2FA(db))
	auth.POST("/verify-password", VerifyPassword(db))
	auth.POST("/mnemonic/regenerate", RegenerateMnemonic(db))
	auth.POST("/password", ChangePassword(db))

	api := r.Group("/")
	api.Use(AuthMiddleware(db))
	api.GET("/countries", GetCountries(db))
	api.GET("/payment-methods", GetPaymentMethods(db))
	api.GET("/assets", GetAssets(db))
	api.GET("/client/assets", GetClientAssets(db))
	api.GET("/client/payment-methods", ListClientPaymentMethods(db))
	api.POST("/client/payment-methods", CreateClientPaymentMethod(db))
	api.PUT("/client/payment-methods/:id", UpdateClientPaymentMethod(db))
	api.DELETE("/client/payment-methods/:id", DeleteClientPaymentMethod(db))
	api.GET("/client/wallets", ListClientWallets(db))
	api.POST("/client/wallets", CreateWallet(db))
	api.GET("/client/balances", ListClientBalances(db))
	api.GET("/client/escrows", ListClientEscrows(db))
	api.GET("/client/escrows/:id", GetClientEscrow(db))
	api.GET("/client/transactions/in", ListClientTransactionsIn(db))
	api.GET("/client/transactions/out", ListClientTransactionsOut(db))
	api.GET("/client/transactions/internal", ListClientTransactionsInternal(db))
	api.GET("/client/orders", ListClientOrders(db))
	api.POST("/client/orders", CreateOrder(db))
	api.GET("/orders/:id/messages", ListOrderMessages(db))
	api.POST("/orders/:id/messages", CreateOrderMessage(db, store, cache))
	api.PATCH("/orders/:id/messages/:msgId/read", ReadOrderMessage(db))
	api.PATCH("/notifications/:id/read", ReadNotification(db))

	maxOffers := 1
	api.GET("/offers", ListOffers(db))
	api.GET("/client/offers", ListClientOffers(db))
	api.POST("/client/offers", CreateOffer(db))
	api.PUT("/client/offers/:id", UpdateOffer(db))
	api.POST("/client/offers/:id/enable", EnableOffer(db, maxOffers))
	api.POST("/client/offers/:id/disable", DisableOffer(db))
	api.DELETE("/client/offers/:id", DeleteOffer(db))

	ws := r.Group("/ws")
	ws.Use(AuthMiddleware(db))
	ws.GET("/orders", OrdersWS())
	ws.GET("/orders/:id/chat", OrderChatWS(db, cache))
	ws.GET("/orders/:id/status", OrderStatusWS(db))
	ws.GET("/offers", gin.WrapF(OffersWS()))

	return db, r, ttl
}
