package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"ptop/internal/models"
	"ptop/internal/notifications"
)

func TestNotificationsWS(t *testing.T) {
	db, r, _ := setupTest(t)
	notifications.SetDB(db)
	r.GET("/ws/notifications", AuthMiddleware(db), NotificationsWS(db))
	srv := httptest.NewServer(r)
	defer srv.Close()

	w := httptest.NewRecorder()
	body := `{"username":"user","password":"pass","password_confirm":"pass"}`
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"user","password":"pass"}`
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

	var client models.Client
	if err := db.Where("username = ?", "user").First(&client).Error; err != nil {
		t.Fatalf("query client: %v", err)
	}

	n1 := models.Notification{ClientID: client.ID, Type: "test1"}
	if err := db.Create(&n1).Error; err != nil {
		t.Fatalf("create n1: %v", err)
	}

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/notifications?token=" + tok.AccessToken
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status %d", resp.StatusCode)
	}
	defer conn.Close()

	var recv models.Notification
	if err := conn.ReadJSON(&recv); err != nil {
		t.Fatalf("read initial: %v", err)
	}
	if recv.ID != n1.ID {
		t.Fatalf("unexpected initial %s", recv.ID)
	}

	n2 := models.Notification{ClientID: client.ID, Type: "test2"}
	if err := db.Create(&n2).Error; err != nil {
		t.Fatalf("create n2: %v", err)
	}
	notifications.Broadcast(client.ID, n2)

	if err := conn.ReadJSON(&recv); err != nil {
		t.Fatalf("read broadcast: %v", err)
	}
	if recv.ID != n2.ID {
		t.Fatalf("unexpected broadcast %s", recv.ID)
	}
}
