package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ptop/internal/models"
)

func TestClientPaymentMethods(t *testing.T) {
	db, r, _ := setupTest(t)
	country := models.Country{Name: "Russia"}
	method := models.PaymentMethod{Name: "Bank"}
	db.Create(&country)
	db.Create(&method)

	// register first user
	body := `{"username":"cpmuser1","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register1 status %d", w.Code)
	}
	var reg1 registerResp
	if err := json.Unmarshal(w.Body.Bytes(), &reg1); err != nil {
		t.Fatalf("reg1 parse: %v", err)
	}
	token1 := reg1.AccessToken

	// create client payment method
	createBody := fmt.Sprintf(`{"country_id":"%s","payment_method_id":"%s","city":"Moscow","post_code":"101000","name":"Main"}`, country.ID, method.ID)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/payment-methods", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token1)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create status %d", w.Code)
	}
	var created models.ClientPaymentMethod
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("create parse: %v", err)
	}

	// create with same name should fail
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/payment-methods", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token1)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate status %d", w.Code)
	}

	// list should return one
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/client/payment-methods", nil)
	req.Header.Set("Authorization", "Bearer "+token1)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d", w.Code)
	}
	var list []models.ClientPaymentMethod
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("list parse: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list length %d", len(list))
	}

	// register second user
	body = `{"username":"cpmuser2","password":"pass","password_confirm":"pass"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register2 status %d", w.Code)
	}
	var reg2 registerResp
	if err := json.Unmarshal(w.Body.Bytes(), &reg2); err != nil {
		t.Fatalf("reg2 parse: %v", err)
	}
	token2 := reg2.AccessToken

	// attempt delete by other user should fail
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/client/payment-methods/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("delete foreign status %d", w.Code)
	}

	// delete by owner
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/client/payment-methods/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token1)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status %d", w.Code)
	}
}
