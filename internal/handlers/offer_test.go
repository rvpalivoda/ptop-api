package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ptop/internal/models"
)

func TestOfferLifecycle(t *testing.T) {
	db, r, _ := setupTest(t)

	// register and login
	body := `{"username":"offuser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"offuser","password":"pass"}`
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

	// create assets, payment method, client payment method
	asset1 := models.Asset{Name: "USD_offer", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_offer", Type: models.AssetTypeCrypto, IsActive: true}
	country := models.Country{Name: "CountryOffer"}
	method := models.PaymentMethod{
		Name:         "BankOffer",
		MethodGroup:  "bank_transfer",
		IsRealtime:   false,
		FeeSide:      models.FeeSideSender,
		KycLevelHint: models.KycLevelHintLow,
	}
	if err := db.Create(&asset1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&asset2).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&country).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&method).Error; err != nil {
		t.Fatal(err)
	}

	var client models.Client
	db.Where("username = ?", "offuser").First(&client)
	cpm := models.ClientPaymentMethod{ClientID: client.ID, CountryID: country.ID, PaymentMethodID: method.ID, Name: "pm1"}
	if err := db.Create(&cpm).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM offers")
		db.Exec("DELETE FROM client_payment_methods")
		db.Exec("DELETE FROM payment_methods")
		db.Exec("DELETE FROM assets")
		db.Exec("DELETE FROM countries")
	})

	// create first offer
	reqBody := OfferRequest{
		MaxAmount:              "100",
		MinAmount:              "10",
		Amount:                 "50",
		Price:                  "0.12345678",
		FromAssetID:            asset1.ID,
		ToAssetID:              asset2.ID,
		OrderExpirationTimeout: 20,
		ClientPaymentMethodIDs: []string{cpm.ID},
	}
	b, _ := json.Marshal(reqBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create offer status %d", w.Code)
	}
	var offer1 models.Offer
	json.Unmarshal(w.Body.Bytes(), &offer1)
	if offer1.IsEnabled {
		t.Fatalf("offer should be disabled by default")
	}
	if offer1.Price.String() != "0.12345678" {
		t.Fatalf("price not set")
	}

	// enable first offer
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+offer1.ID+"/enable", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("enable offer status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &offer1)
	if !offer1.IsEnabled || offer1.EnabledAt == nil {
		t.Fatalf("offer not enabled")
	}
	if offer1.TTL.Before(time.Now().Add(29 * 24 * time.Hour)) {
		t.Fatalf("ttl not updated")
	}

	// create second offer
	reqBody.Conditions = "second"
	b, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create offer2 status %d", w.Code)
	}
	var offer2 models.Offer
	json.Unmarshal(w.Body.Bytes(), &offer2)

	// try enable second offer should fail due to limit 1
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+offer2.ID+"/enable", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("enable second should fail")
	}

	// list active offers - should be 1
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/offers", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	var list []models.Offer
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].ID != offer1.ID {
		t.Fatalf("active offers list unexpected")
	}

	// list client disabled offers should include offer2
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/offers?enabled=false", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) == 0 {
		t.Fatalf("disabled offers not returned")
	}

	// disable first offer
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+offer1.ID+"/disable", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("disable offer status %d", w.Code)
	}

	// enable second offer now should succeed
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+offer2.ID+"/enable", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("enable second status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &offer2)

	// update second offer
	reqBody.Conditions = "updated"
	reqBody.Price = "0.87654321"
	b, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/client/offers/"+offer2.ID, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update offer status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &offer2)
	if offer2.Conditions != "updated" {
		t.Fatalf("offer not updated")
	}
	if offer2.Price.String() != "0.87654321" {
		t.Fatalf("price not updated")
	}

	// disable second offer
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+offer2.ID+"/disable", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("disable second status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &offer2)
	if offer2.IsEnabled {
		t.Fatalf("offer not disabled")
	}
}
