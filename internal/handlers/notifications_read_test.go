package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ptop/internal/models"
)

func TestReadNotification(t *testing.T) {
	db, r, _ := setupTest(t)

	// register first user
	w := httptest.NewRecorder()
	body := `{"username":"user1","password":"pass","password_confirm":"pass"}`
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"user1","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("user1 login status %d", w.Code)
	}
	var tok1 struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tok1)

	// register second user
	w = httptest.NewRecorder()
	body = `{"username":"user2","password":"pass","password_confirm":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	body = `{"username":"user2","password":"pass"}`
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("user2 login status %d", w.Code)
	}
	var tok2 struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tok2)

	var client1 models.Client
	db.Where("username = ?", "user1").First(&client1)

	// create notification for user1
	n := models.Notification{ClientID: client1.ID, Type: "test"}
	if err := db.Create(&n).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}

	// user1 marks notification read
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/notifications/"+n.ID+"/read", nil)
	req.Header.Set("Authorization", "Bearer "+tok1.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("read status %d", w.Code)
	}
	var upd models.Notification
	json.Unmarshal(w.Body.Bytes(), &upd)
	if upd.ReadAt == nil {
		t.Fatalf("expected readAt not nil")
	}

	// user2 tries to mark someone else's notification
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/notifications/"+n.ID+"/read", nil)
	req.Header.Set("Authorization", "Bearer "+tok2.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
