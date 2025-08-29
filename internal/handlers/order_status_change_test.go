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

func TestOrderStatusChangeFlow(t *testing.T) {
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

	// create assets and offer
	asset1 := models.Asset{Name: "USD_status", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_status", Type: models.AssetTypeCrypto, IsActive: true}
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

	// buyer creates order
	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order create %d", w.Code)
	}
	var ord models.Order
	json.Unmarshal(w.Body.Bytes(), &ord)

	// open WS for both sides
	wsURLBuyer := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/orders/" + ord.ID + "/status?token=" + buyerTok.AccessToken
	connBuyer, resp, err := websocket.DefaultDialer.Dial(wsURLBuyer, nil)
	if err != nil || resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("buyer ws: %v %d", err, resp.StatusCode)
	}
	defer connBuyer.Close()
	wsURLSeller := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/orders/" + ord.ID + "/status?token=" + sellerTok.AccessToken
	connSeller, resp2, err := websocket.DefaultDialer.Dial(wsURLSeller, nil)
	if err != nil || resp2.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("seller ws: %v %d", err, resp2.StatusCode)
	}
	defer connSeller.Close()

	// buyer marks PAID
	paidAt := time.Now().UTC().Format(time.RFC3339)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/paid", bytes.NewBufferString("{\"paidAt\":\""+paidAt+"\"}"))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("paid %d", w.Code)
	}

	var sevt OrderStatusEvent
	if err := connBuyer.ReadJSON(&sevt); err != nil {
		t.Fatalf("buyer status evt: %v", err)
	}
	if sevt.Type != "order.status_changed" || sevt.Order.Status != models.OrderStatusPaid {
		t.Fatalf("unexpected paid evt buyer: %#v", sevt)
	}
	if err := connSeller.ReadJSON(&sevt); err != nil {
		t.Fatalf("seller status evt: %v", err)
	}
	if sevt.Order.Status != models.OrderStatusPaid {
		t.Fatalf("unexpected paid evt seller: %#v", sevt)
	}

	// seller releases
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/release", nil)
	req.Header.Set("Authorization", "Bearer "+sellerTok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("release %d", w.Code)
	}
	if err := connBuyer.ReadJSON(&sevt); err != nil {
		t.Fatalf("buyer release evt: %v", err)
	}
	if sevt.Order.Status != models.OrderStatusReleased {
		t.Fatalf("unexpected release evt buyer: %#v", sevt)
	}
	if err := connSeller.ReadJSON(&sevt); err != nil {
		t.Fatalf("seller release evt: %v", err)
	}
	if sevt.Order.Status != models.OrderStatusReleased {
		t.Fatalf("unexpected release evt seller: %#v", sevt)
	}
}

func TestResolveDisputeFlow(t *testing.T) {
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

	// register arbiter
	w = httptest.NewRecorder()
	body = `{"username":"arb","password":"pass","password_confirm":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	w = httptest.NewRecorder()
	body = `{"username":"arb","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("arb login status %d", w.Code)
	}
	var arbTok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &arbTok)

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

	// create assets and offer
	asset1 := models.Asset{Name: "USD_resolve", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_resolve", Type: models.AssetTypeCrypto, IsActive: true}
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

	// buyer creates order
	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order create %d", w.Code)
	}
	var ord models.Order
	json.Unmarshal(w.Body.Bytes(), &ord)

	// open WS for both sides
	wsURLBuyer := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/orders/" + ord.ID + "/status?token=" + buyerTok.AccessToken
	connBuyer, resp, err := websocket.DefaultDialer.Dial(wsURLBuyer, nil)
	if err != nil || resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("buyer ws: %v %d", err, resp.StatusCode)
	}
	defer connBuyer.Close()
	wsURLSeller := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/orders/" + ord.ID + "/status?token=" + sellerTok.AccessToken
	connSeller, resp2, err := websocket.DefaultDialer.Dial(wsURLSeller, nil)
	if err != nil || resp2.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("seller ws: %v %d", err, resp2.StatusCode)
	}
	defer connSeller.Close()

	// buyer marks PAID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/paid", bytes.NewBufferString("{}"))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("paid %d", w.Code)
	}
	var evt OrderStatusEvent
	if err := connBuyer.ReadJSON(&evt); err != nil {
		t.Fatalf("buyer paid evt: %v", err)
	}
	if evt.Order.Status != models.OrderStatusPaid {
		t.Fatalf("unexpected paid evt buyer: %#v", evt)
	}
	if err := connSeller.ReadJSON(&evt); err != nil {
		t.Fatalf("seller paid evt: %v", err)
	}
	if evt.Order.Status != models.OrderStatusPaid {
		t.Fatalf("unexpected paid evt seller: %#v", evt)
	}

	// seller opens dispute
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/dispute", bytes.NewBufferString("{\"reason\":\"bad\"}"))
	req.Header.Set("Authorization", "Bearer "+sellerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("dispute %d", w.Code)
	}
	if err := connBuyer.ReadJSON(&evt); err != nil {
		t.Fatalf("buyer dispute evt: %v", err)
	}
	if evt.Order.Status != models.OrderStatusDispute {
		t.Fatalf("unexpected dispute evt buyer: %#v", evt)
	}
	if err := connSeller.ReadJSON(&evt); err != nil {
		t.Fatalf("seller dispute evt: %v", err)
	}
	if evt.Order.Status != models.OrderStatusDispute {
		t.Fatalf("unexpected dispute evt seller: %#v", evt)
	}

	// arbiter resolves to CANCELLED
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/dispute/resolve", bytes.NewBufferString("{\"result\":\"CANCELLED\",\"comment\":\"arb\"}"))
	req.Header.Set("Authorization", "Bearer "+arbTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("resolve %d", w.Code)
	}
	var ofull models.OrderFull
	json.Unmarshal(w.Body.Bytes(), &ofull)
	if ofull.Status != models.OrderStatusCancelled || ofull.CancelReason == nil || *ofull.CancelReason != "arb" {
		t.Fatalf("unexpected resolve response %#v", ofull)
	}
	if err := connBuyer.ReadJSON(&evt); err != nil {
		t.Fatalf("buyer resolve evt: %v", err)
	}
	if evt.Order.Status != models.OrderStatusCancelled {
		t.Fatalf("unexpected resolve evt buyer: %#v", evt)
	}
	if err := connSeller.ReadJSON(&evt); err != nil {
		t.Fatalf("seller resolve evt: %v", err)
	}
	if evt.Order.Status != models.OrderStatusCancelled {
		t.Fatalf("unexpected resolve evt seller: %#v", evt)
	}

	// second order for RELEASED result
	w = httptest.NewRecorder()
	body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234"}`
	req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order2 create %d", w.Code)
	}
	var ord2 models.Order
	json.Unmarshal(w.Body.Bytes(), &ord2)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord2.ID+"/paid", bytes.NewBufferString("{}"))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order2 paid %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord2.ID+"/dispute", bytes.NewBufferString("{}"))
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order2 dispute %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord2.ID+"/dispute/resolve", bytes.NewBufferString("{\"result\":\"RELEASED\"}"))
	req.Header.Set("Authorization", "Bearer "+arbTok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("order2 resolve %d", w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &ofull)
	if ofull.Status != models.OrderStatusReleased || ofull.ReleasedAt == nil {
		t.Fatalf("unexpected order2 resolve %#v", ofull)
	}
}
