package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/net/websocket"
	"ptop/internal/models"
	"ptop/internal/orderchat"
)

func TestOrderChatWS(t *testing.T) {
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

	asset1 := models.Asset{Name: "USD_ws", Type: models.AssetTypeFiat, IsActive: true}
	asset2 := models.Asset{Name: "BTC_ws", Type: models.AssetTypeCrypto, IsActive: true}
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

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/orders/" + ord.ID + "/chat"

	// hacker tries to connect
	cfg, _ := websocket.NewConfig(wsURL, "http://example.com")
	cfg.Header = http.Header{"Authorization": {"Bearer " + hackerTok.AccessToken}}
	_, err := websocket.DialConfig(cfg)
	if err == nil {
		t.Fatalf("expected forbidden, got nil")
	}

	// buyer connects and sends message
	cfg, _ = websocket.NewConfig(wsURL, "http://example.com")
	cfg.Header = http.Header{"Authorization": {"Bearer " + buyerTok.AccessToken}}
	buyerConn, err := websocket.DialConfig(cfg)
	if err != nil {
		t.Fatalf("buyer dial: %v", err)
	}
	if err := websocket.JSON.Send(buyerConn, OrderMessageRequest{Content: "hello"}); err != nil {
		t.Fatalf("write: %v", err)
	}
	var echo orderchat.Event
	if err := websocket.JSON.Receive(buyerConn, &echo); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if echo.Type != string(models.MessageTypeText) || echo.Message.Content != "hello" {
		t.Fatalf("unexpected echo %#v", echo)
	}
	buyerConn.Close()

	// seller connects after message and receives history
	cfg, _ = websocket.NewConfig(wsURL, "http://example.com")
	cfg.Header = http.Header{"Authorization": {"Bearer " + sellerTok.AccessToken}}
	sellerConn, err := websocket.DialConfig(cfg)
	if err != nil {
		t.Fatalf("seller dial: %v", err)
	}
	defer sellerConn.Close()
	var history orderchat.Event
	if err := websocket.JSON.Receive(sellerConn, &history); err != nil {
		t.Fatalf("history read: %v", err)
	}
	if history.Type != string(models.MessageTypeText) || history.Message.Content != "hello" {
		t.Fatalf("unexpected content %#v", history)
	}
	var dbMsg models.OrderMessage
	if err := db.Where("id = ?", history.Message.ID).First(&dbMsg).Error; err != nil {
		t.Fatalf("db message: %v", err)
	}

	// buyer uploads file via REST, seller should receive it over WS
	fileBody := &bytes.Buffer{}
	mw := multipart.NewWriter(fileBody)
	fw, err := mw.CreateFormFile("file", "test.png")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	fw.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	mw.Close()

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/messages", fileBody)
	req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("file msg status %d", w.Code)
	}

	var fileEvt orderchat.Event
	if err := websocket.JSON.Receive(sellerConn, &fileEvt); err != nil {
		t.Fatalf("file read: %v", err)
	}
	if fileEvt.Type != string(models.MessageTypeFile) || fileEvt.Message.FileURL == nil {
		t.Fatalf("unexpected file event %#v", fileEvt)
	}
	if !strings.HasPrefix(*fileEvt.Message.FileURL, "https://example.com/") {
		t.Fatalf("unexpected file url %s", *fileEvt.Message.FileURL)
	}
}
