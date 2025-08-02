package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ptop/internal/models"
)

func TestBalanceHandler(t *testing.T) {
	db, r, _ := setupTest(t)

	body := `{"username":"baluser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"baluser","password":"pass"}`
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

	asset := models.Asset{Name: "BTC_balance", Type: "crypto"}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}

	w = httptest.NewRecorder()
	body = `{"asset_id":"` + asset.ID + `"}`
	req, _ = http.NewRequest("POST", "/client/wallets", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("wallet status %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/balances", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("balances status %d", w.Code)
	}
	var list []models.Balance
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 balance, got %d", len(list))
	}
	if !list[0].Amount.IsZero() {
		t.Fatalf("amount not zero")
	}
}
