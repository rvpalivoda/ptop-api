package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ptop/internal/models"
)

func TestReferenceHandlers(t *testing.T) {
	db, r, _ := setupTest(t)
	// seed data
	country := models.Country{Name: "Russia"}
	method := models.PaymentMethod{Name: "Bank"}
	activeAsset := models.Asset{Name: "Ruble", Type: models.AssetTypeFiat, IsActive: true}
	inactiveAsset := models.Asset{Name: "Inactive", Type: models.AssetTypeFiat}
	db.Create(&country)
	db.Create(&method)
	db.Create(&activeAsset)
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
	if len(methods) != 1 || methods[0].Name != "Bank" {
		t.Fatalf("methods data %+v", methods)
	}

	// assets
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
	if len(assets) != 1 || assets[0].Name != "Ruble" {
		t.Fatalf("assets data %+v", assets)
	}
}
