package notifications

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ptop/internal/models"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&models.Notification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	SetDB(db)
	return db
}

func dialWS(t *testing.T, ch chan<- models.Notification) *websocket.Conn {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		var n models.Notification
		if err := conn.ReadJSON(&n); err == nil {
			ch <- n
		}
	}))
	t.Cleanup(server.Close)
	u := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

func TestSend(t *testing.T) {
	db := setupDB(t)
	n := models.Notification{ClientID: "c1", Type: "test", LinkTo: "/link"}
	if err := db.Create(&n).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	ch := make(chan models.Notification, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		defer conn.Close()
		var recv models.Notification
		if err := conn.ReadJSON(&recv); err != nil {
			t.Fatalf("read: %v", err)
		}
		ch <- recv
	}))
	defer server.Close()

	u := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if err := Send(conn, n); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case recv := <-ch:
		if recv.ID != n.ID {
			t.Fatalf("expected %s, got %s", n.ID, recv.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	var updated models.Notification
	if err := db.First(&updated, "id = ?", n.ID).Error; err != nil {
		t.Fatalf("query: %v", err)
	}
	if updated.SentAt == nil {
		t.Fatalf("SentAt not updated")
	}
}

func TestBroadcast(t *testing.T) {
	db := setupDB(t)
	clientID := "c1"
	n := models.Notification{ClientID: clientID, Type: "test", LinkTo: "/link"}
	if err := db.Create(&n).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	ch1 := make(chan models.Notification, 1)
	conn1 := dialWS(t, ch1)
	defer conn1.Close()
	AddClient(clientID, conn1)
	defer RemoveClient(clientID, conn1)

	ch2 := make(chan models.Notification, 1)
	conn2 := dialWS(t, ch2)
	defer conn2.Close()
	AddClient(clientID, conn2)
	defer RemoveClient(clientID, conn2)

	Broadcast(clientID, n)

	for i, ch := range []chan models.Notification{ch1, ch2} {
		select {
		case recv := <-ch:
			if recv.ID != n.ID {
				t.Fatalf("conn %d: expected %s, got %s", i+1, n.ID, recv.ID)
			}
		case <-time.After(time.Second):
			t.Fatalf("conn %d: timeout", i+1)
		}
	}

	var updated models.Notification
	if err := db.First(&updated, "id = ?", n.ID).Error; err != nil {
		t.Fatalf("query: %v", err)
	}
	if updated.SentAt == nil {
		t.Fatalf("SentAt not updated")
	}
}
