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

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"ptop/internal/models"
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
	header := http.Header{"Authorization": {"Bearer " + hackerTok.AccessToken}}
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %v %v", err, resp)
	}

	// buyer connects and sends message
	header = http.Header{"Authorization": {"Bearer " + buyerTok.AccessToken}}
	buyerConn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("buyer dial: %v", err)
	}
	if err := buyerConn.WriteJSON(OrderMessageRequest{Content: "hello"}); err != nil {
		t.Fatalf("write: %v", err)
	}
	var echo orderChatEvent
	if err := buyerConn.ReadJSON(&echo); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if echo.Type != string(models.MessageTypeText) || echo.Message.Content != "hello" {
		t.Fatalf("unexpected echo %#v", echo)
	}
	buyerConn.Close()

	// seller connects after message and receives history
	header = http.Header{"Authorization": {"Bearer " + sellerTok.AccessToken}}
	sellerConn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("seller dial: %v", err)
	}
	defer sellerConn.Close()
	var history orderChatEvent
	if err := sellerConn.ReadJSON(&history); err != nil {
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

	var fileEvt orderChatEvent
	if err := sellerConn.ReadJSON(&fileEvt); err != nil {
		t.Fatalf("file read: %v", err)
	}
	if fileEvt.Type != string(models.MessageTypeFile) || fileEvt.Message.FileURL == nil {
		t.Fatalf("unexpected file event %#v", fileEvt)
	}
	if !strings.HasPrefix(*fileEvt.Message.FileURL, "https://example.com/") {
		t.Fatalf("unexpected file url %s", *fileEvt.Message.FileURL)
	}
}
