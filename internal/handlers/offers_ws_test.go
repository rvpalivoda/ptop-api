package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"ptop/internal/models"
)

type offerWSEvent struct {
	Type  string           `json:"type"`
	Offer models.OfferFull `json:"offer"`
}

func TestOffersWS(t *testing.T) {
	db, r, _ := setupTest(t)
	srv := httptest.NewServer(r)
	defer srv.Close()

	// register and login user
	w := httptest.NewRecorder()
	body := `{"username":"seller","password":"pass","password_confirm":"pass"}`
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"seller","password":"pass"}`
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
	asset1 := models.Asset{Name: "USD_ws", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_ws", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset1).Error; err != nil {
		t.Fatalf("asset1: %v", err)
	}
	if err := db.Create(&asset2).Error; err != nil {
		t.Fatalf("asset2: %v", err)
	}

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/offers"
	header := http.Header{"Authorization": {"Bearer " + tok.AccessToken}}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// create offer
	w = httptest.NewRecorder()
	body = fmt.Sprintf(`{"max_amount":"100","min_amount":"1","amount":"50","price":"0.1","type":"sell","from_asset_id":"%s","to_asset_id":"%s","order_expiration_timeout":20}`, asset1.ID, asset2.ID)
	req, _ = http.NewRequest("POST", "/client/offers", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create status %d", w.Code)
	}
	var created models.Offer
	json.Unmarshal(w.Body.Bytes(), &created)

	var evt offerWSEvent
	if err := conn.ReadJSON(&evt); err != nil {
		t.Fatalf("read create: %v", err)
	}
	if evt.Type != "created" || evt.Offer.ID != created.ID || evt.Offer.FromAsset.ID != asset1.ID {
		t.Fatalf("unexpected create event %#v", evt)
	}

	// update offer
	w = httptest.NewRecorder()
	body = fmt.Sprintf(`{"max_amount":"100","min_amount":"1","amount":"50","price":"0.2","type":"sell","from_asset_id":"%s","to_asset_id":"%s","order_expiration_timeout":20}`, asset1.ID, asset2.ID)
	req, _ = http.NewRequest("PUT", "/client/offers/"+created.ID, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update status %d", w.Code)
	}
	if err := conn.ReadJSON(&evt); err != nil {
		t.Fatalf("read update: %v", err)
	}
	if evt.Type != "updated" || evt.Offer.Price.Cmp(decimal.RequireFromString("0.2")) != 0 {
		t.Fatalf("unexpected update event %#v", evt)
	}

	// delete offer
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/client/offers/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status %d", w.Code)
	}
	if err := conn.ReadJSON(&evt); err != nil {
		t.Fatalf("read delete: %v", err)
	}
	if evt.Type != "deleted" || evt.Offer.ID != created.ID {
		t.Fatalf("unexpected delete event %#v", evt)
	}
}
