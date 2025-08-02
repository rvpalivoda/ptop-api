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

	asset := models.Asset{Name: "BTC_escrow", Type: "crypto"}
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
