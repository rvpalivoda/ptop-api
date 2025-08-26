package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/offers?token=" + tok.AccessToken
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status %d", resp.StatusCode)
	}
	defer conn.Close()
	if err := conn.WriteJSON(struct{}{}); err != nil {
		t.Fatalf("init write: %v", err)
	}

	// create offer (inactive)
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

	// enable offer -> should broadcast
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+created.ID+"/enable", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("enable status %d", w.Code)
	}
	if err := conn.ReadJSON(&evt); err != nil {
		t.Fatalf("read enable: %v", err)
	}
	if evt.Type != "created" || evt.Offer.ID != created.ID {
		t.Fatalf("unexpected enable event %#v", evt)
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

	// disable offer -> should broadcast deleted
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/client/offers/"+created.ID+"/disable", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("disable status %d", w.Code)
	}
	if err := conn.ReadJSON(&evt); err != nil {
		t.Fatalf("read disable: %v", err)
	}
	if evt.Type != "deleted" || evt.Offer.ID != created.ID {
		t.Fatalf("unexpected disable event %#v", evt)
	}

	// delete offer (already inactive) -> no event
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/client/offers/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status %d", w.Code)
	}
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err := conn.ReadJSON(&evt); err == nil {
		t.Fatalf("unexpected event after delete %#v", evt)
	}
}
