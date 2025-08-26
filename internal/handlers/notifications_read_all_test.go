package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ptop/internal/models"
)

func TestReadAllNotifications(t *testing.T) {
	db, r, _ := setupTest(t)

	// register user
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
		t.Fatalf("login status %d", w.Code)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &tok)

	var client models.Client
	db.Where("username = ?", "user1").First(&client)

	// create notifications
	n1 := models.Notification{ClientID: client.ID, Type: "test", LinkTo: "/l1"}
	n2 := models.Notification{ClientID: client.ID, Type: "test", LinkTo: "/l2"}
	db.Create(&n1)
	db.Create(&n2)
	db.Create(&models.Notification{ClientID: "other", Type: "test", LinkTo: "/l3"})

	// mark all read
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/notifications/read-all", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("read-all status %d", w.Code)
	}
	var resp NotificationsReadAllResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Count != 2 {
		t.Fatalf("expected 2, got %d", resp.Count)
	}

	var updated1, updated2 models.Notification
	db.First(&updated1, "id = ?", n1.ID)
	db.First(&updated2, "id = ?", n2.ID)
	if updated1.ReadAt == nil || updated2.ReadAt == nil {
		t.Fatalf("expected notifications read")
	}
}
