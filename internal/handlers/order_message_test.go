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

func TestOrderMessageHandler(t *testing.T) {
	db, r, _ := setupTest(t)

	// register seller
	body := `{"username":"seller","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	body = `{"username":"seller","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("seller login status %d", w.Code)
	}
	var sellerTok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &sellerTok)
	var seller models.Client
	db.Where("username = ?", "seller").First(&seller)

	// register buyer
	w = httptest.NewRecorder()
	body = `{"username":"buyer","password":"pass","password_confirm":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	body = `{"username":"buyer","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buyer login status %d", w.Code)
	}
	var buyerTok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &buyerTok)
	var buyer models.Client
	db.Where("username = ?", "buyer").First(&buyer)

	// register hacker
	w = httptest.NewRecorder()
	body = `{"username":"hacker","password":"pass","password_confirm":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	body = `{"username":"hacker","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("hacker login status %d", w.Code)
	}
	var hackerTok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &hackerTok)

	asset1 := models.Asset{Name: "USD_msg", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_msg", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset1).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}
	if err := db.Create(&asset2).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}
	offer := models.Offer{
		MaxAmount:              decimal.RequireFromString("100"),
		MinAmount:              decimal.RequireFromString("1"),
		Amount:                 decimal.RequireFromString("50"),
		Price:                  decimal.RequireFromString("0.1"),
		FromAssetID:            asset1.ID,
		ToAssetID:              asset2.ID,
		OrderExpirationTimeout: 10,
		TTL:                    time.Now().Add(24 * time.Hour),
		ClientID:               seller.ID,
	}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatalf("offer: %v", err)
	}

	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order status %d", w.Code)
	}
	var ord models.Order
	json.Unmarshal(w.Body.Bytes(), &ord)

	// buyer sends message
	w = httptest.NewRecorder()
	body = `{"content":"hello"}`
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/messages", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create message status %d", w.Code)
	}
	var msg models.OrderMessage
	json.Unmarshal(w.Body.Bytes(), &msg)

	// hacker tries to get messages
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/"+ord.ID+"/messages", nil)
	req.Header.Set("Authorization", "Bearer "+hackerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", w.Code)
	}

	// buyer lists messages
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/"+ord.ID+"/messages", nil)
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list messages status %d", w.Code)
	}
	var list []models.OrderMessage
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 message, got %d", len(list))
	}

	// seller marks message read
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/orders/"+ord.ID+"/messages/"+msg.ID+"/read", nil)
	req.Header.Set("Authorization", "Bearer "+sellerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("read status %d", w.Code)
	}
	var upd models.OrderMessage
	json.Unmarshal(w.Body.Bytes(), &upd)
	if upd.ReadAt == nil {
		t.Fatalf("expected read at not nil")
	}
}
