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

	// register and login seller
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
		t.Fatalf("login seller status %d", w.Code)
	}
	var tokSeller struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tokSeller)

	// set seller pincode
	w = httptest.NewRecorder()
	body = `{"password":"pass","pincode":"1234"}`
	req, _ = http.NewRequest("POST", "/auth/pincode", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tokSeller.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("seller pincode status %d", w.Code)
	}

	// register and login buyer
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
		t.Fatalf("login buyer status %d", w.Code)
	}
	var tokBuyer struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tokBuyer)

	// set buyer pincode
	w = httptest.NewRecorder()
	body = `{"password":"pass","pincode":"1234"}`
	req, _ = http.NewRequest("POST", "/auth/pincode", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tokBuyer.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buyer pincode status %d", w.Code)
	}

	var seller models.Client
	db.Where("username = ?", "seller").First(&seller)
	var buyer models.Client
	db.Where("username = ?", "buyer").First(&buyer)

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
	cpmSeller := models.ClientPaymentMethod{
		ClientID:        seller.ID,
		CountryID:       country.ID,
		PaymentMethodID: method.ID,
		City:            "Moscow",
		PostCode:        "101000",
		Name:            "Seller",
	}
	if err := db.Create(&cpmSeller).Error; err != nil {
		t.Fatalf("cpm seller: %v", err)
	}
	cpmBuyer := models.ClientPaymentMethod{
		ClientID:        buyer.ID,
		CountryID:       country.ID,
		PaymentMethodID: method.ID,
		City:            "Moscow",
		PostCode:        "101000",
		Name:            "Buyer",
	}
	if err := db.Create(&cpmBuyer).Error; err != nil {
		t.Fatalf("cpm buyer: %v", err)
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

	// attempt to create order for own offer
	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234","client_payment_method_id":"` + cpmSeller.ID + `"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tokSeller.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", w.Code)
	}

	// create order by buyer
	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234","client_payment_method_id":"` + cpmBuyer.ID + `"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tokBuyer.AccessToken)
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
	req.Header.Set("Authorization", "Bearer "+tokBuyer.AccessToken)
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
	if list[0].Buyer.ID != buyer.ID || list[0].Seller.ID != seller.ID {
		t.Fatalf("unexpected client ids")
	}
	if list[0].Author.ID != buyer.ID || list[0].OfferOwner.ID != seller.ID {
		t.Fatalf("unexpected author or offer owner")
	}
	if list[0].ClientPaymentMethod == nil || list[0].ClientPaymentMethod.ID != cpmBuyer.ID {
		t.Fatalf("missing client payment method")
	}
	if list[0].ClientPaymentMethod.Country.ID != country.ID || list[0].ClientPaymentMethod.PaymentMethod.ID != method.ID {
		t.Fatalf("missing nested payment method data")
	}

	// create second order for pagination test
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tokBuyer.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("second order status %d", w.Code)
	}

	// test pagination
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/orders?limit=1", nil)
	req.Header.Set("Authorization", "Bearer "+tokBuyer.AccessToken)
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
	req.Header.Set("Authorization", "Bearer "+tokBuyer.AccessToken)
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
	req.Header.Set("Authorization", "Bearer "+tokSeller.AccessToken)
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
