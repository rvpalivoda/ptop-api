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

func TestOrderAutoExpireCancelsAndBroadcasts(t *testing.T) {
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
    var sellerTok struct{ AccessToken string `json:"access_token"` }
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
    var buyerTok struct{ AccessToken string `json:"access_token"` }
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
    asset1 := models.Asset{Name: "USD_exp", Type: models.AssetTypeFiat, IsActive: true}
    asset2 := models.Asset{Name: "BTC_exp", Type: models.AssetTypeCrypto, IsActive: true}
    if err := db.Create(&asset1).Error; err != nil { t.Fatalf("asset1: %v", err) }
    if err := db.Create(&asset2).Error; err != nil { t.Fatalf("asset2: %v", err) }
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
    if err := db.Create(&offer).Error; err != nil { t.Fatalf("offer: %v", err) }

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

    // Make it expired
    if err := db.Model(&models.Order{}).Where("id = ?", ord.ID).
        Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
        t.Fatalf("set expired: %v", err)
    }

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

    // Start expirer with short interval
    exp := NewOrderExpirer(db, 50*time.Millisecond)
    exp.Start()
    defer exp.Stop()

    // Expect CANCELLED event on both connections
    var evt OrderStatusEvent
    connBuyer.SetReadDeadline(time.Now().Add(2 * time.Second))
    if err := connBuyer.ReadJSON(&evt); err != nil {
        t.Fatalf("buyer status evt: %v", err)
    }
    if evt.Order.Status != models.OrderStatusCancelled {
        t.Fatalf("unexpected buyer status: %s", evt.Order.Status)
    }
    connSeller.SetReadDeadline(time.Now().Add(2 * time.Second))
    if err := connSeller.ReadJSON(&evt); err != nil {
        t.Fatalf("seller status evt: %v", err)
    }
    if evt.Order.Status != models.OrderStatusCancelled {
        t.Fatalf("unexpected seller status: %s", evt.Order.Status)
    }
}

func TestOrderAutoExpirePaidGoesToDisputeAndBroadcasts(t *testing.T) {
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
    if w.Code != http.StatusOK { t.Fatalf("seller login status %d", w.Code) }
    var sellerTok struct{ AccessToken string `json:"access_token"` }
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
    if w.Code != http.StatusOK { t.Fatalf("buyer login status %d", w.Code) }
    var buyerTok struct{ AccessToken string `json:"access_token"` }
    json.Unmarshal(w.Body.Bytes(), &buyerTok)

    // set pincode for buyer
    w = httptest.NewRecorder()
    body = `{"password":"pass","pincode":"1234"}`
    req, _ = http.NewRequest("POST", "/auth/pincode", bytes.NewBufferString(body))
    req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("buyer pincode status %d", w.Code) }

    // create assets and offer
    asset1 := models.Asset{Name: "USD_exp2", Type: models.AssetTypeFiat, IsActive: true}
    asset2 := models.Asset{Name: "BTC_exp2", Type: models.AssetTypeCrypto, IsActive: true}
    if err := db.Create(&asset1).Error; err != nil { t.Fatalf("asset1: %v", err) }
    if err := db.Create(&asset2).Error; err != nil { t.Fatalf("asset2: %v", err) }
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
    if err := db.Create(&offer).Error; err != nil { t.Fatalf("offer: %v", err) }

    // buyer creates order
    w = httptest.NewRecorder()
    body = `{"offer_id":"` + offer.ID + `","amount":"5","pin_code":"1234"}`
    req, _ = http.NewRequest("POST", "/client/orders", bytes.NewBufferString(body))
    req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("order create %d", w.Code) }
    var ord models.Order
    json.Unmarshal(w.Body.Bytes(), &ord)

    // mark PAID
    paidAt := time.Now().UTC().Format(time.RFC3339)
    w = httptest.NewRecorder()
    req, _ = http.NewRequest("POST", "/orders/"+ord.ID+"/paid", bytes.NewBufferString("{\"paidAt\":\""+paidAt+"\"}"))
    req.Header.Set("Authorization", "Bearer "+buyerTok.AccessToken)
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("paid %d", w.Code) }

    // expire it
    if err := db.Model(&models.Order{}).Where("id = ?", ord.ID).
        Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
        t.Fatalf("set expired: %v", err)
    }

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

    // Start expirer with short interval
    exp := NewOrderExpirer(db, 50*time.Millisecond)
    exp.Start()
    defer exp.Stop()

    // Expect DISPUTE event on both connections
    var evt OrderStatusEvent
    connBuyer.SetReadDeadline(time.Now().Add(2 * time.Second))
    if err := connBuyer.ReadJSON(&evt); err != nil {
        t.Fatalf("buyer status evt: %v", err)
    }
    if evt.Order.Status != models.OrderStatusDispute {
        t.Fatalf("unexpected buyer status: %s", evt.Order.Status)
    }
    connSeller.SetReadDeadline(time.Now().Add(2 * time.Second))
    if err := connSeller.ReadJSON(&evt); err != nil {
        t.Fatalf("seller status evt: %v", err)
    }
    if evt.Order.Status != models.OrderStatusDispute {
        t.Fatalf("unexpected seller status: %s", evt.Order.Status)
    }
}
