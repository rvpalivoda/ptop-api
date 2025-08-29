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

func TestGetOrderActions(t *testing.T) {
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

	// set pincode for buyer
	w = httptest.NewRecorder()
	body = `{"password":"pass","pincode":"1234"}`
	req, _ = http.NewRequest("POST", "/auth/pincode", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buyer pincode status %d", w.Code)
	}

	// create assets and offer
	asset1 := models.Asset{Name: "USD_actions", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_actions", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset1).Error; err != nil {
		t.Fatalf("asset1: %v", err)
	}
	if err := db.Create(&asset2).Error; err != nil {
		t.Fatalf("asset2: %v", err)
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

	// buyer creates order
	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order create %d", w.Code)
	}
	var ord models.Order
	json.Unmarshal(w.Body.Bytes(), &ord)

	// actions for WAIT_PAYMENT
	var act struct {
		Actions []string `json:"actions"`
	}
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/"+ord.ID+"/actions", nil)
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buyer actions status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &act)
	if len(act.Actions) != 2 || act.Actions[0] != "markPaid" || act.Actions[1] != "cancel" {
		t.Fatalf("buyer actions unexpected %#v", act)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/"+ord.ID+"/actions", nil)
	req.Header.Set("Authorization", "Bearer "+sellerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("seller actions status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &act)
	if len(act.Actions) != 1 || act.Actions[0] != "cancel" {
		t.Fatalf("seller actions unexpected %#v", act)
	}

	// buyer marks paid
	w = httptest.NewRecorder()
	paidAt := time.Now().UTC().Format(time.RFC3339)
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/paid", bytes.NewBufferString(`{"paidAt":"`+paidAt+`"}`))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("paid %d", w.Code)
	}

	// actions for PAID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/"+ord.ID+"/actions", nil)
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buyer actions paid status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &act)
	if len(act.Actions) != 1 || act.Actions[0] != "dispute" {
		t.Fatalf("buyer actions paid unexpected %#v", act)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/"+ord.ID+"/actions", nil)
	req.Header.Set("Authorization", "Bearer "+sellerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("seller actions paid status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &act)
	if len(act.Actions) != 2 || act.Actions[0] != "release" || act.Actions[1] != "dispute" {
		t.Fatalf("seller actions paid unexpected %#v", act)
	}

	// seller releases
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/release", nil)
	req.Header.Set("Authorization", "Bearer "+sellerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("release %d", w.Code)
	}

	// after release no actions
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/"+ord.ID+"/actions", nil)
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buyer actions released status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &act)
	if len(act.Actions) != 0 {
		t.Fatalf("buyer actions after release unexpected %#v", act)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/"+ord.ID+"/actions", nil)
	req.Header.Set("Authorization", "Bearer "+sellerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("seller actions released status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &act)
	if len(act.Actions) != 0 {
		t.Fatalf("seller actions after release unexpected %#v", act)
	}
}
