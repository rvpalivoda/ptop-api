package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"ptop/internal/models"
)

func TestOrderStatusWS(t *testing.T) {
	db, r, _ := setupTest(t)
	srv := httptest.NewServer(r)
	defer srv.Close()

	// register seller
	body := `{"username":"seller","password":"pass","password_confirm":"pass"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	body = `{"username":"seller","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("seller login status %d", w.Code)
	}
	var sellerTok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &sellerTok)
	var seller models.Client
	db.Where("username = ?", "seller").First(&seller)

	// register buyer
	w = httptest.NewRecorder()
	body = `{"username":"buyer","password":"pass","password_confirm":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	body = `{"username":"buyer","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buyer login status %d", w.Code)
	}
	var buyerTok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &buyerTok)
	var buyer models.Client
	db.Where("username = ?", "buyer").First(&buyer)

	// set pincode for buyer
	w = httptest.NewRecorder()
	body = `{"password":"pass","pincode":"1234"}`
	req, _ = http.NewRequest("POST", "/auth/pincode", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buyer pincode status %d", w.Code)
	}

	// register hacker
	w = httptest.NewRecorder()
	body = `{"username":"hacker","password":"pass","password_confirm":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	body = `{"username":"hacker","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("hacker login status %d", w.Code)
	}
	var hackerTok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &hackerTok)

	asset1 := models.Asset{Name: "USD_ws_status", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_ws_status", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset1).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}
	if err := db.Create(&asset2).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}
	offer := models.Offer{
		MaxAmount:              decimal.RequireFromString("100"),
		MinAmount:              decimal.RequireFromString("1"),
		Amount:                 decimal.RequireFromString("50"),
		Price:                  decimal.RequireFromString("0.1"),
		FromAssetID:            asset1.ID,
		ToAssetID:              asset2.ID,
		OrderExpirationTimeout: 10,
		TTL:                    time.Now().Add(24 * time.Hour),
		ClientID:               seller.ID,
	}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatalf("offer: %v", err)
	}

	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order status %d", w.Code)
	}
	var ord models.Order
	json.Unmarshal(w.Body.Bytes(), &ord)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/orders/" + ord.ID + "/status"

	// hacker tries to connect
	header := http.Header{"Authorization": {"Bearer " + hackerTok.AccessToken}}
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %v %v", err, resp)
	}

	header = http.Header{"Authorization": {"Bearer " + buyerTok.AccessToken}}
	buyerConn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("buyer dial: %v", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("buyer handshake status %d", resp.StatusCode)
	}
	defer buyerConn.Close()

	header = http.Header{"Authorization": {"Bearer " + sellerTok.AccessToken}}
	sellerConn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("seller dial: %v", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("seller handshake status %d", resp.StatusCode)
	}
	defer sellerConn.Close()

	if err := db.Model(&ord).Update("status", models.OrderStatusPaid).Error; err != nil {
		t.Fatalf("update status: %v", err)
	}
	var full models.Order
	if err := db.Preload("Offer").
		Preload("Buyer").
		Preload("Seller").
		Preload("Author").
		Preload("OfferOwner").
		Preload("FromAsset").
		Preload("ToAsset").
		Where("id = ?", ord.ID).First(&full).Error; err != nil {
		t.Fatalf("preload: %v", err)
	}
	broadcastOrderStatus(full)

	var evt orderStatusEvent
	if err := buyerConn.ReadJSON(&evt); err != nil {
		t.Fatalf("buyer read: %v", err)
	}
	if evt.Type != "order.status_changed" || evt.Order.Status != models.OrderStatusPaid {
		t.Fatalf("unexpected buyer event %#v", evt)
	}
	if err := sellerConn.ReadJSON(&evt); err != nil {
		t.Fatalf("seller read: %v", err)
	}
	if evt.Type != "order.status_changed" || evt.Order.Status != models.OrderStatusPaid {
		t.Fatalf("unexpected seller event %#v", evt)
	}
}
