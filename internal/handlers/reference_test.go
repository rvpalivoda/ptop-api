package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"ptop/internal/models"
)

func TestReferenceHandlers(t *testing.T) {
	db, r, _ := setupTest(t)
	// seed data
	country := models.Country{Name: "Russia"}
	method := models.PaymentMethod{
		Name:         "Bank",
		MethodGroup:  "bank_transfer",
		IsRealtime:   false,
		FeeSide:      models.FeeSideSender,
		KycLevelHint: models.KycLevelHintLow,
	}
	activeAsset := models.Asset{Name: "Ruble", Type: models.AssetTypeFiat, IsActive: true}
	cryptoAsset := models.Asset{Name: "BTC", Type: models.AssetTypeCrypto, IsActive: true}
	inactiveAsset := models.Asset{Name: "Inactive", Type: models.AssetTypeFiat}
	db.Create(&country)
	db.Create(&method)
	db.Create(&activeAsset)
	db.Create(&cryptoAsset)
	db.Create(&inactiveAsset)

	// register user
	body := `{"username":"refuser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register status %d", w.Code)
	}
	var reg registerResp
	if err := json.Unmarshal(w.Body.Bytes(), &reg); err != nil {
		t.Fatalf("register parse: %v", err)
	}
	token := reg.AccessToken

	var client models.Client
	if err := db.Where("username = ?", "refuser").First(&client).Error; err != nil {
		t.Fatalf("find client: %v", err)
	}
	wallet := models.Wallet{ClientID: client.ID, AssetID: cryptoAsset.ID, Value: "addr", DerivationIndex: 1, IsEnabled: true}
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatalf("wallet: %v", err)
	}
	balance := models.Balance{ClientID: client.ID, AssetID: cryptoAsset.ID, Amount: decimal.NewFromInt(10), AmountEscrow: decimal.Zero}
	if err := db.Create(&balance).Error; err != nil {
		t.Fatalf("balance: %v", err)
	}

	// countries
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/countries", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("countries status %d", w.Code)
	}
	var countries []models.Country
	if err := json.Unmarshal(w.Body.Bytes(), &countries); err != nil {
		t.Fatalf("countries parse: %v", err)
	}
	if len(countries) != 1 || countries[0].Name != "Russia" {
		t.Fatalf("countries data %+v", countries)
	}

	// payment methods
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/payment-methods", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("payment methods status %d", w.Code)
	}
	var methods []models.PaymentMethod
	if err := json.Unmarshal(w.Body.Bytes(), &methods); err != nil {
		t.Fatalf("methods parse: %v", err)
	}
	if len(methods) != 1 || methods[0].Name != "Bank" || methods[0].MethodGroup != "bank_transfer" || methods[0].IsRealtime {
		t.Fatalf("methods data %+v", methods)
	}

	// assets without wallets
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/assets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("assets status %d", w.Code)
	}
	var assets []models.Asset
	if err := json.Unmarshal(w.Body.Bytes(), &assets); err != nil {
		t.Fatalf("assets parse: %v", err)
	}
	if len(assets) != 2 {
		t.Fatalf("assets length %d", len(assets))
	}
	names := map[string]bool{}
	for _, a := range assets {
		if !a.IsActive {
			t.Fatalf("inactive asset returned: %+v", a)
		}
		names[a.Name] = true
	}
	if !names["Ruble"] || !names["BTC"] {
		t.Fatalf("assets names %+v", names)
	}

	// client assets with wallets
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/assets", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("client assets status %d", w.Code)
	}
	var clientAssets []AssetWithWallet
	if err := json.Unmarshal(w.Body.Bytes(), &clientAssets); err != nil {
		t.Fatalf("client assets parse: %v", err)
	}
	if len(clientAssets) != 1 {
		t.Fatalf("client assets length %d", len(clientAssets))
	}
	ca := clientAssets[0]
	if ca.Name != "BTC" || ca.Value != "addr" || !ca.Amount.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("client asset btc %+v", ca)
	}
}
