package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"ptop/internal/models"
)

func TestEscrowHandler(t *testing.T) {
	db, r, _ := setupTest(t)

	body := `{"username":"escuser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"escuser","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status %d", w.Code)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tok)

	var client models.Client
	db.Where("username = ?", "escuser").First(&client)

	asset := models.Asset{Name: "BTC_escrow", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}
	esc := models.Escrow{ClientID: client.ID, AssetID: asset.ID, Amount: decimal.RequireFromString("1")}
	if err := db.Create(&esc).Error; err != nil {
		t.Fatalf("escrow: %v", err)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/escrows", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("escrows status %d", w.Code)
	}
	var list []models.Escrow
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 escrow, got %d", len(list))
	}
	if list[0].Amount.Cmp(decimal.RequireFromString("1")) != 0 {
		t.Fatalf("amount mismatch")
	}
}

func TestGetEscrowHandler(t *testing.T) {
	db, r, _ := setupTest(t)

	body := `{"username":"escuser2","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"escuser2","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status %d", w.Code)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tok)

	var client models.Client
	db.Where("username = ?", "escuser2").First(&client)

	asset := models.Asset{Name: "BTC_escrow2", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}

	offer := models.Offer{
		MaxAmount:              decimal.RequireFromString("10"),
		MinAmount:              decimal.RequireFromString("1"),
		Amount:                 decimal.RequireFromString("5"),
		Price:                  decimal.RequireFromString("1"),
		Type:                   models.OfferTypeBuy,
		FromAssetID:            asset.ID,
		ToAssetID:              asset.ID,
		OrderExpirationTimeout: 15,
		TTL:                    time.Now().Add(time.Hour),
		ClientID:               client.ID,
	}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatalf("offer: %v", err)
	}

	order := models.Order{
		OfferID:     offer.ID,
		BuyerID:     client.ID,
		SellerID:    client.ID,
		FromAssetID: asset.ID,
		ToAssetID:   asset.ID,
		Amount:      decimal.RequireFromString("1"),
		Price:       decimal.RequireFromString("1"),
		Status:      models.OrderStatusPaid,
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("order: %v", err)
	}

	esc1 := models.Escrow{ClientID: client.ID, AssetID: asset.ID, Amount: decimal.RequireFromString("1"), OfferID: &offer.ID, OrderID: &order.ID}
	if err := db.Create(&esc1).Error; err != nil {
		t.Fatalf("escrow1: %v", err)
	}
	esc2 := models.Escrow{ClientID: client.ID, AssetID: asset.ID, Amount: decimal.RequireFromString("2")}
	if err := db.Create(&esc2).Error; err != nil {
		t.Fatalf("escrow2: %v", err)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/escrows/"+esc1.ID, nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get escrow status %d", w.Code)
	}
	var resp struct {
		Escrow struct {
			ID     string        `json:"id"`
			Client models.Client `json:"client"`
			Asset  models.Asset  `json:"asset"`
			Offer  *models.Offer `json:"offer"`
			Order  *models.Order `json:"order"`
		} `json:"escrow"`
		NextID string `json:"nextId"`
		PrevID string `json:"prevId"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Escrow.ID != esc1.ID || resp.NextID != esc2.ID || resp.PrevID != "" {
		t.Fatalf("unexpected nav ids")
	}
	if resp.Escrow.Client.ID != client.ID || resp.Escrow.Asset.ID != asset.ID {
		t.Fatalf("missing related data")
	}
	if resp.Escrow.Offer == nil || resp.Escrow.Order == nil {
		t.Fatalf("expected offer and order")
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/escrows/"+esc1.ID+"?dir=next", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("next escrow status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Escrow.ID != esc2.ID || resp.PrevID != esc1.ID {
		t.Fatalf("next navigation failed")
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/escrows/"+esc2.ID+"?dir=prev", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("prev escrow status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Escrow.ID != esc1.ID {
		t.Fatalf("prev navigation failed")
	}
}
