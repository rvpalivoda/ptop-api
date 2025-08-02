package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ptop/internal/models"
)

type registerResp struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	Mnemonic     []MnemonicWord `json:"mnemonic"`
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func TestRegisterLoginRefresh(t *testing.T) {
	_, r, _ := setupTest(t)

	// register
	body := `{"username":"user1","password":"pass","password_confirm":"pass"}`
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
	if len(reg.Mnemonic) != 12 {
		t.Fatalf("mnemonic length %d", len(reg.Mnemonic))
	}

	// login
	body = `{"username":"user1","password":"pass"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status %d", w.Code)
	}
	var log tokenResp
	if err := json.Unmarshal(w.Body.Bytes(), &log); err != nil {
		t.Fatalf("login parse: %v", err)
	}

	// refresh
	body = `{"refresh_token":"` + log.RefreshToken + `"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("refresh status %d", w.Code)
	}
}

func TestRecoverFlow(t *testing.T) {
	_, r, _ := setupTest(t)
	// register new user
	body := `{"username":"recuser","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register status %d", w.Code)
	}
	var reg registerResp
	if err := json.Unmarshal(w.Body.Bytes(), &reg); err != nil {
		t.Fatalf("parse reg: %v", err)
	}

	// recover challenge
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/auth/recover/recuser", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("challenge status %d", w.Code)
	}
	var ch RecoverChallengeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &ch); err != nil {
		t.Fatalf("challenge parse: %v", err)
	}
	if len(ch.Positions) != 3 {
		t.Fatalf("positions %v", ch.Positions)
	}

	// recover with first three words
	phrases := []RecoverPhrase{
		{Position: 1, Word: reg.Mnemonic[0].Word},
		{Position: 2, Word: reg.Mnemonic[1].Word},
		{Position: 3, Word: reg.Mnemonic[2].Word},
	}
	reqBody, _ := json.Marshal(RecoverRequest{Username: "recuser", Phrases: phrases})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/recover", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("recover status %d", w.Code)
	}
}

func TestProfileAndSettings(t *testing.T) {
	db, r, _ := setupTest(t)
	// register
	body := `{"username":"user2","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var reg registerResp
	json.Unmarshal(w.Body.Bytes(), &reg)

	// login
	body = `{"username":"user2","password":"pass"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var log tokenResp
	json.Unmarshal(w.Body.Bytes(), &log)

	// profile
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+log.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("profile status %d", w.Code)
	}

	// change username
	body = `{"password":"pass","new_username":"newname"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/username", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+log.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("change username %d", w.Code)
	}

	// set pincode
	body = `{"password":"pass","pincode":"1234"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/pincode", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+log.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("pincode status %d", w.Code)
	}

	// change password
	body = `{"old_password":"pass","new_password":"pass2","confirm_password":"pass2"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+log.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("change password %d", w.Code)
	}

	// login with new password
	body = `{"username":"newname","password":"pass2"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login new password %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &log)

	// enable 2fa
	body = `{"password":"pass2"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/2fa/enable", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+log.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("enable2fa %d", w.Code)
	}

	// check profile flags
	var client models.Client
	if err := db.Where("username = ?", "newname").First(&client).Error; err != nil {
		t.Fatalf("db lookup: %v", err)
	}
	if !client.TwoFAEnabled || client.PinCode == nil {
		t.Fatalf("settings not saved")
	}
}
