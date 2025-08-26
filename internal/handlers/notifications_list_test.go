package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ptop/internal/models"
)

func TestListNotifications(t *testing.T) {
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

	var client1, client2 models.Client
	db.Where("username = ?", "user1").First(&client1)
	db.Where("username = ?", "user2").First(&client2)

	// create notifications for user1
	var n1, n2, n3 models.Notification
	n1 = models.Notification{ClientID: client1.ID, Type: "test"}
	db.Create(&n1)
	n2 = models.Notification{ClientID: client1.ID, Type: "test"}
	db.Create(&n2)
	n3 = models.Notification{ClientID: client1.ID, Type: "test"}
	db.Create(&n3)
	// create notification for user2
	db.Create(&models.Notification{ClientID: client2.ID, Type: "test"})

	// user1 list first page
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/notifications?limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+tok1.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d", w.Code)
	}
	var list []models.Notification
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
	if list[0].ID != n3.ID || list[1].ID != n2.ID {
		t.Fatalf("unexpected order")
	}

	// user1 list second page
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/notifications?limit=2&offset=2", nil)
	req.Header.Set("Authorization", "Bearer "+tok1.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list2 status %d", w.Code)
	}
	list = nil
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].ID != n1.ID {
		t.Fatalf("expected n1")
	}

	// user2 list
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+tok2.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("user2 list status %d", w.Code)
	}
	list = nil
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].ClientID != client2.ID {
		t.Fatalf("expected only user2 notifications")
	}
}
