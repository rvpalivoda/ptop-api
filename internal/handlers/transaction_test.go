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

func TestTransactionHandlers(t *testing.T) {
	db, r, _ := setupTest(t)

	body := `{"username":"tranuser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"tranuser","password":"pass"}`
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
	if err := db.Where("username = ?", "tranuser").First(&client).Error; err != nil {
		t.Fatalf("client: %v", err)
	}

	asset := models.Asset{Name: "BTC_trx", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}

	wal := models.Wallet{ClientID: client.ID, AssetID: asset.ID, Value: "addr", IsEnabled: true, EnabledAt: time.Now()}
	if err := db.Create(&wal).Error; err != nil {
		t.Fatalf("wallet: %v", err)
	}

	tInOld := models.TransactionIn{ClientID: client.ID, WalletID: wal.ID, AssetID: asset.ID, Amount: decimal.NewFromInt(1), Status: models.TransactionInStatusConfirmed, CreatedAt: time.Now().Add(-time.Minute)}
	tInNew := models.TransactionIn{ClientID: client.ID, WalletID: wal.ID, AssetID: asset.ID, Amount: decimal.NewFromInt(2), Status: models.TransactionInStatusPending, CreatedAt: time.Now()}
	if err := db.Create(&tInOld).Error; err != nil {
		t.Fatalf("txin1: %v", err)
	}
	if err := db.Create(&tInNew).Error; err != nil {
		t.Fatalf("txin2: %v", err)
	}

	tOutOld := models.TransactionOut{ClientID: client.ID, AssetID: asset.ID, Amount: decimal.NewFromInt(3), Status: models.TransactionOutStatusPending, CreatedAt: time.Now().Add(-time.Minute)}
	tOutNew := models.TransactionOut{ClientID: client.ID, AssetID: asset.ID, Amount: decimal.NewFromInt(4), Status: models.TransactionOutStatusConfirmed, CreatedAt: time.Now()}
	if err := db.Create(&tOutOld).Error; err != nil {
		t.Fatalf("txout1: %v", err)
	}
	if err := db.Create(&tOutNew).Error; err != nil {
		t.Fatalf("txout2: %v", err)
	}

	other := models.Client{Username: "other"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("other: %v", err)
	}

	tIntOld := models.TransactionInternal{AssetID: asset.ID, Amount: decimal.NewFromInt(5), FromClientID: client.ID, ToClientID: other.ID, Status: models.TransactionInternalStatusConfirmed, CreatedAt: time.Now().Add(-time.Minute)}
	tIntNew := models.TransactionInternal{AssetID: asset.ID, Amount: decimal.NewFromInt(6), FromClientID: other.ID, ToClientID: client.ID, Status: models.TransactionInternalStatusProcessing, CreatedAt: time.Now()}
	tIntOther := models.TransactionInternal{AssetID: asset.ID, Amount: decimal.NewFromInt(7), FromClientID: other.ID, ToClientID: other.ID, Status: models.TransactionInternalStatusFailed, CreatedAt: time.Now()}
	if err := db.Create(&tIntOld).Error; err != nil {
		t.Fatalf("txint1: %v", err)
	}
	if err := db.Create(&tIntNew).Error; err != nil {
		t.Fatalf("txint2: %v", err)
	}
	if err := db.Create(&tIntOther).Error; err != nil {
		t.Fatalf("txint3: %v", err)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/transactions/in?limit=1&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("in list status %d", w.Code)
	}
	var inList []models.TransactionIn
	json.Unmarshal(w.Body.Bytes(), &inList)
	if len(inList) != 1 {
		t.Fatalf("expected 1, got %d", len(inList))
	}
	if inList[0].ID != tInNew.ID {
		t.Fatalf("expected newest tx")
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/transactions/in?limit=1&offset=1", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("in list offset status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &inList)
	if len(inList) != 1 || inList[0].ID != tInOld.ID {
		t.Fatalf("pagination failed")
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/transactions/out?limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("out list status %d", w.Code)
	}
	var outList []models.TransactionOut
	json.Unmarshal(w.Body.Bytes(), &outList)
	if len(outList) != 2 {
		t.Fatalf("expected 2, got %d", len(outList))
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/transactions/internal?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("internal list status %d", w.Code)
	}
	var intList []models.TransactionInternal
	json.Unmarshal(w.Body.Bytes(), &intList)
	if len(intList) != 2 {
		t.Fatalf("expected 2, got %d", len(intList))
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/transactions/internal?limit=1&offset=1", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("internal list offset status %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &intList)
	if len(intList) != 1 || intList[0].ID != tIntOld.ID {
		t.Fatalf("internal pagination failed")
	}
}
