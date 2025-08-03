package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"

	"ptop/internal/models"
)

func TestWalletHandlers(t *testing.T) {
	db, r, _ := setupTest(t)

	// register and login
	body := `{"username":"waluser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"waluser","password":"pass"}`
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

	// create assets
	seed := bytes.Repeat([]byte{0x01}, 32)
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("master: %v", err)
	}
	xpub, err := master.Neuter()
	if err != nil {
		t.Fatalf("neuter: %v", err)
	}
	crypto := models.Asset{Name: "BTC_wallet", Type: models.AssetTypeCrypto, Xpub: xpub.String(), IsActive: true}
	fiat := models.Asset{Name: "USD_wallet", Type: models.AssetTypeFiat, IsActive: true}
	if err := db.Create(&crypto).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}
	if err := db.Create(&fiat).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}

	// create wallet
	w = httptest.NewRecorder()
	body = `{"asset_id":"` + crypto.ID + `"}`
	req, _ = http.NewRequest("POST", "/client/wallets", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create status %d", w.Code)
	}
	var wal models.Wallet
	json.Unmarshal(w.Body.Bytes(), &wal)
	if wal.Value == "" {
		t.Fatalf("empty value")
	}

	// duplicate active wallet
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/wallets", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("dup status %d", w.Code)
	}

	// non-crypto asset
	w = httptest.NewRecorder()
	body = `{"asset_id":"` + fiat.ID + `"}`
	req, _ = http.NewRequest("POST", "/client/wallets", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("fiat status %d", w.Code)
	}

	// disable wallet manually
	now := time.Now()
	db.Model(&models.Wallet{}).Where("id = ?", wal.ID).Updates(map[string]any{"is_enabled": false, "disabled_at": now})

	// create wallet again
	w = httptest.NewRecorder()
	body = `{"asset_id":"` + crypto.ID + `"}`
	req, _ = http.NewRequest("POST", "/client/wallets", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("second create status %d", w.Code)
	}

	// list wallets
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/wallets", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d", w.Code)
	}
	var list []models.Wallet
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 wallet, got %d", len(list))
	}
	if list[0].ID == wal.ID {
		t.Fatalf("disabled wallet returned")
	}
}

func TestCreateWalletFake(t *testing.T) {
	t.Setenv("DEBUG_FAKE_NETWORK", "true")
	db, r, _ := setupTest(t)

	body := `{"username":"fakeuser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"fakeuser","password":"pass"}`
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

	asset := models.Asset{Name: "BTC_fake", Type: models.AssetTypeCrypto, IsActive: true}
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
		t.Fatalf("create status %d", w.Code)
	}
	var wal models.Wallet
	json.Unmarshal(w.Body.Bytes(), &wal)
	if wal.Value == "" {
		t.Fatalf("empty value")
	}
	var client models.Client
	if err := db.Where("username = ?", "fakeuser").First(&client).Error; err != nil {
		t.Fatalf("client: %v", err)
	}
	exp := fmt.Sprintf("fake:%s:%s:%d", asset.ID, client.ID, 0)
	if wal.Value != exp {
		t.Fatalf("expected %s, got %s", exp, wal.Value)
	}
}
