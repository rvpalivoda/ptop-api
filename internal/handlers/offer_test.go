package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
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
		Type:                   models.OfferTypeBuy,
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
	var list []models.OfferFull
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].ID != offer1.ID {
		t.Fatalf("active offers list unexpected")
	}
	if list[0].FromAsset.ID == "" || list[0].Client.ID == "" {
		t.Fatalf("nested data not returned")
	}
	if list[0].Client.Rating.Cmp(decimal.Zero) != 0 || list[0].Client.OrdersCount != 0 {
		t.Fatalf("client fields not default")
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
	if list[0].Client.ID == "" || len(list[0].ClientPaymentMethods) == 0 {
		t.Fatalf("client nested data missing")
	}
	if list[0].Client.Rating.Cmp(decimal.Zero) != 0 || list[0].Client.OrdersCount != 0 {
		t.Fatalf("client fields not default in client offers")
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

func TestListOffersFilters(t *testing.T) {
	db, r, _ := setupTest(t)

	// register and login first user
	body := `{"username":"user1","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"user1","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login1 status %d", w.Code)
	}
	var tok1 struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tok1)

	// register and login second user
	w = httptest.NewRecorder()
	body = `{"username":"user2","password":"pass","password_confirm":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"user2","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login2 status %d", w.Code)
	}
	var tok2 struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tok2)

	// prepare assets and payment methods
	usd := models.Asset{Name: "USD_f", Type: models.AssetTypeFiat, IsActive: true}
	btc := models.Asset{Name: "BTC_c", Type: models.AssetTypeCrypto, IsActive: true}
	country := models.Country{Name: "CountryF"}
	pm1 := models.PaymentMethod{Name: "Bank1", MethodGroup: "bank", IsRealtime: false, FeeSide: models.FeeSideSender, KycLevelHint: models.KycLevelHintLow}
	pm2 := models.PaymentMethod{Name: "Bank2", MethodGroup: "bank", IsRealtime: false, FeeSide: models.FeeSideSender, KycLevelHint: models.KycLevelHintLow}
	if err := db.Create(&usd).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&btc).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&country).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&pm1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&pm2).Error; err != nil {
		t.Fatal(err)
	}

	// client payment methods for each user
	var c1, c2 models.Client
	db.Where("username = ?", "user1").First(&c1)
	db.Where("username = ?", "user2").First(&c2)
	cpm1 := models.ClientPaymentMethod{ClientID: c1.ID, CountryID: country.ID, PaymentMethodID: pm1.ID, Name: "pm1"}
	cpm2 := models.ClientPaymentMethod{ClientID: c2.ID, CountryID: country.ID, PaymentMethodID: pm2.ID, Name: "pm2"}
	if err := db.Create(&cpm1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&cpm2).Error; err != nil {
		t.Fatal(err)
	}

	// offer1: buy crypto (USD -> BTC)
	reqBody := OfferRequest{
		MaxAmount:              "100",
		MinAmount:              "10",
		Amount:                 "50",
		Price:                  "1",
		Type:                   models.OfferTypeBuy,
		FromAssetID:            usd.ID,
		ToAssetID:              btc.ID,
		ClientPaymentMethodIDs: []string{cpm1.ID},
	}
	b, _ := json.Marshal(reqBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok1.AccessToken)
	r.ServeHTTP(w, req)
	var off1 models.Offer
	json.Unmarshal(w.Body.Bytes(), &off1)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+off1.ID+"/enable", nil)
	req.Header.Set("Authorization", "Bearer "+tok1.AccessToken)
	r.ServeHTTP(w, req)

	// offer2: sell crypto (BTC -> USD)
	reqBody = OfferRequest{
		MaxAmount:              "500",
		MinAmount:              "200",
		Amount:                 "300",
		Price:                  "1",
		Type:                   models.OfferTypeSell,
		FromAssetID:            btc.ID,
		ToAssetID:              usd.ID,
		ClientPaymentMethodIDs: []string{cpm2.ID},
	}
	b, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok2.AccessToken)
	r.ServeHTTP(w, req)
	var off2 models.Offer
	json.Unmarshal(w.Body.Bytes(), &off2)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+off2.ID+"/enable", nil)
	req.Header.Set("Authorization", "Bearer "+tok2.AccessToken)
	r.ServeHTTP(w, req)

	// filter: buy offers with amount range and payment method
	w = httptest.NewRecorder()
	url := "/offers?from_asset=" + usd.ID + "&to_asset=" + btc.ID + "&min_amount=20&max_amount=80&payment_method=" + pm1.ID + "&type=" + models.OfferTypeBuy
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tok1.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d", w.Code)
	}
	var list []models.OfferFull
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].ID != off1.ID {
		t.Fatalf("unexpected filter result")
	}
}

func TestListOffersPagination(t *testing.T) {
	db, r, _ := setupTest(t)

	tokens := make([]string, 3)
	for i := 1; i <= 3; i++ {
		// register
		body := fmt.Sprintf(`{"username":"u%d","password":"pass","password_confirm":"pass"}`, i)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// login
		body = fmt.Sprintf(`{"username":"u%d","password":"pass"}`, i)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("login%d status %d", i, w.Code)
		}
		var tok struct {
			AccessToken string `json:"access_token"`
		}
		json.Unmarshal(w.Body.Bytes(), &tok)
		tokens[i-1] = tok.AccessToken
	}

	usd := models.Asset{Name: "USDp", Type: models.AssetTypeFiat, IsActive: true}
	btc := models.Asset{Name: "BTCp", Type: models.AssetTypeCrypto, IsActive: true}
	country := models.Country{Name: "CountryP"}
	pm := models.PaymentMethod{Name: "BankP", MethodGroup: "bank", IsRealtime: false, FeeSide: models.FeeSideSender, KycLevelHint: models.KycLevelHintLow}
	if err := db.Create(&usd).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&btc).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&country).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&pm).Error; err != nil {
		t.Fatal(err)
	}

	var offers []models.Offer
	for i := 1; i <= 3; i++ {
		var c models.Client
		db.Where("username = ?", fmt.Sprintf("u%d", i)).First(&c)
		cpm := models.ClientPaymentMethod{ClientID: c.ID, CountryID: country.ID, PaymentMethodID: pm.ID, Name: fmt.Sprintf("pm%d", i)}
		if err := db.Create(&cpm).Error; err != nil {
			t.Fatal(err)
		}

		reqBody := OfferRequest{
			MaxAmount:              "100",
			MinAmount:              "10",
			Amount:                 "50",
			Price:                  "1",
			Type:                   models.OfferTypeBuy,
			FromAssetID:            usd.ID,
			ToAssetID:              btc.ID,
			ClientPaymentMethodIDs: []string{cpm.ID},
		}
		b, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/client/offers", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokens[i-1])
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create offer%d status %d", i, w.Code)
		}
		var off models.Offer
		json.Unmarshal(w.Body.Bytes(), &off)
		offers = append(offers, off)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/client/offers/"+off.ID+"/enable", nil)
		req.Header.Set("Authorization", "Bearer "+tokens[i-1])
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("enable offer%d status %d", i, w.Code)
		}
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/offers?limit=2&offset=1", nil)
	req.Header.Set("Authorization", "Bearer "+tokens[0])
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d", w.Code)
	}
	var list []models.OfferFull
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 2 || list[0].ID != offers[1].ID || list[1].ID != offers[0].ID {
		t.Fatalf("unexpected pagination result")
	}
}
