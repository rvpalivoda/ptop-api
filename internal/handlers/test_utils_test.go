package handlers

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// setupTest создаёт in-memory БД и маршруты для тестов.
func setupTest(t *testing.T) (*gorm.DB, *gin.Engine, map[string]time.Duration) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ttl := map[string]time.Duration{"access": time.Minute, "refresh": time.Hour}

	r := gin.Default()
	auth := r.Group("/auth")
	auth.POST("/register", Register(db, ttl))
	auth.POST("/login", Login(db, ttl))
	auth.POST("/refresh", Refresh(db, ttl))
	auth.GET("/recover/:username", RecoverChallenge(db))
	auth.POST("/recover", Recover(db, ttl))
	auth.Use(AuthMiddleware(db))
	auth.GET("/profile", Profile(db))
	auth.POST("/username", ChangeUsername(db))
	auth.POST("/pincode", SetPinCode(db))
	auth.POST("/2fa/enable", Enable2FA(db))
	auth.POST("/password", ChangePassword(db))

	api := r.Group("/")
	api.Use(AuthMiddleware(db))
	api.GET("/countries", GetCountries(db))
	api.GET("/payment-methods", GetPaymentMethods(db))
	api.GET("/assets", GetAssets(db))
	api.GET("/client/payment-methods", ListClientPaymentMethods(db))
	api.POST("/client/payment-methods", CreateClientPaymentMethod(db))
	api.DELETE("/client/payment-methods/:id", DeleteClientPaymentMethod(db))
	api.GET("/client/wallets", ListClientWallets(db))
	api.POST("/client/wallets", CreateWallet(db))

	maxOffers := 1
	api.GET("/offers", ListOffers(db))
	api.GET("/client/offers", ListClientOffers(db))
	api.POST("/client/offers", CreateOffer(db))
	api.PUT("/client/offers/:id", UpdateOffer(db))
	api.POST("/client/offers/:id/enable", EnableOffer(db, maxOffers))
	api.POST("/client/offers/:id/disable", DisableOffer(db))

	return db, r, ttl
}
