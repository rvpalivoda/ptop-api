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

func TestOrderHandler(t *testing.T) {
	db, r, _ := setupTest(t)

	body := `{"username":"orduser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"orduser","password":"pass"}`
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

	// set pincode
	w = httptest.NewRecorder()
	body = `{"password":"pass","pincode":"1234"}`
	req, _ = http.NewRequest("POST", "/auth/pincode", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("pincode status %d", w.Code)
	}

	var client models.Client
	db.Where("username = ?", "orduser").First(&client)

	asset1 := models.Asset{Name: "USD_order", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_order", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset1).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}
	if err := db.Create(&asset2).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}

	country := models.Country{Name: "Russia"}
	method := models.PaymentMethod{
		Name:         "Bank",
		MethodGroup:  "bank_transfer",
		IsRealtime:   false,
		FeeSide:      models.FeeSideSender,
		KycLevelHint: models.KycLevelHintLow,
	}
	if err := db.Create(&country).Error; err != nil {
		t.Fatalf("country: %v", err)
	}
	if err := db.Create(&method).Error; err != nil {
		t.Fatalf("method: %v", err)
	}
	cpm := models.ClientPaymentMethod{
		ClientID:        client.ID,
		CountryID:       country.ID,
		PaymentMethodID: method.ID,
		City:            "Moscow",
		PostCode:        "101000",
		Name:            "Main",
	}
	if err := db.Create(&cpm).Error; err != nil {
		t.Fatalf("cpm: %v", err)
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
		ClientID:               client.ID,
	}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatalf("offer: %v", err)
	}

	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"0000"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234","client_payment_method_id":"` + cpm.ID + `"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order status %d", w.Code)
	}
	var ord models.Order
	json.Unmarshal(w.Body.Bytes(), &ord)
	if !ord.IsEscrow {
		t.Fatalf("expected escrow true")
	}
	if ord.Status != models.OrderStatusWaitPayment {
		t.Fatalf("unexpected status %s", ord.Status)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/orders", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d", w.Code)
	}
	var list []models.OrderFull
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 order, got %d", len(list))
	}
	if list[0].Offer.ID != offer.ID {
		t.Fatalf("expected offer %s, got %s", offer.ID, list[0].Offer.ID)
	}
	if list[0].FromAsset.ID != asset1.ID || list[0].ToAsset.ID != asset2.ID {
		t.Fatalf("unexpected assets")
	}
	if list[0].Buyer.ID != client.ID || list[0].Seller.ID != client.ID {
		t.Fatalf("unexpected client ids")
	}
	if list[0].Author.ID != client.ID || list[0].OfferOwner.ID != client.ID {
		t.Fatalf("unexpected author or offer owner")
	}
	if list[0].ClientPaymentMethod == nil || list[0].ClientPaymentMethod.ID != cpm.ID {
		t.Fatalf("missing client payment method")
	}
	if list[0].ClientPaymentMethod.Country.ID != country.ID || list[0].ClientPaymentMethod.PaymentMethod.ID != method.ID {
		t.Fatalf("missing nested payment method data")
	}

	// create second order for pagination test
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("second order status %d", w.Code)
	}

	// test pagination
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/orders?limit=1", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("paginated list status %d", w.Code)
	}
	list = nil
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 order with limit, got %d", len(list))
	}

	// test role filter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/orders?role=author", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("author filter status %d", w.Code)
	}
	list = nil
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Fatalf("expected 2 orders for author, got %d", len(list))
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/orders?role=offerOwner", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("owner filter status %d", w.Code)
	}
	list = nil
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Fatalf("expected 2 orders for owner, got %d", len(list))
	}
}
