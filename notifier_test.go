package tgnotify_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tgnotify "github.com/Pebryan354/tg-notify"
)

func TestNew_Validation(t *testing.T) {
	_, err := tgnotify.New("", "123456")
	if err == nil {
		t.Errorf("expected error for empty botToken")
	}

	_, err = tgnotify.New("mytoken", "")
	if err == nil {
		t.Errorf("expected error for empty chatID")
	}

	client, err := tgnotify.New("mytoken", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatalf("expected non-nil client")
	}
}

func TestNotifyError(t *testing.T) {
	var receivedPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/bot12345:TOKEN/sendMessage") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true, "result": {}}`))
	}))
	defer server.Close()

	notifier, err := tgnotify.New("12345:TOKEN", "-100987654321",
		tgnotify.WithBaseURL(server.URL),
		tgnotify.WithAppName("PaymentService"),
		tgnotify.WithEnvironment("production"),
		tgnotify.WithThreadID(42),
	)
	if err != nil {
		t.Fatalf("failed to create notifier: %v", err)
	}

	testErr := errors.New("database connection refused: 127.0.0.1:3306")
	err = notifier.NotifyError(context.Background(), testErr,
		tgnotify.WithExtra(map[string]any{
			"user_id":    1001,
			"request_id": "req-abc-123",
		}),
	)
	if err != nil {
		t.Fatalf("NotifyError failed: %v", err)
	}

	if receivedPayload["chat_id"] != "-100987654321" {
		t.Errorf("expected chat_id -100987654321, got %v", receivedPayload["chat_id"])
	}
	if int(receivedPayload["message_thread_id"].(float64)) != 42 {
		t.Errorf("expected thread_id 42, got %v", receivedPayload["message_thread_id"])
	}

	text := receivedPayload["text"].(string)
	if !strings.Contains(text, "PaymentService") {
		t.Errorf("expected AppName in text, got: %s", text)
	}
	if !strings.Contains(text, "database connection refused") {
		t.Errorf("expected error message in text, got: %s", text)
	}
	if !strings.Contains(text, "req-abc-123") {
		t.Errorf("expected request_id metadata in text, got: %s", text)
	}
	if !strings.Contains(text, "notifier_test.go") {
		t.Errorf("expected caller file in location, got: %s", text)
	}
}

func TestNotifyAlert(t *testing.T) {
	var receivedPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true, "result": {}}`))
	}))
	defer server.Close()

	notifier, err := tgnotify.New("token", "chat123",
		tgnotify.WithBaseURL(server.URL),
		tgnotify.WithAppName("WithdrawEngine"),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = notifier.NotifyAlert(context.Background(),
		"BCA API Timeout",
		"Bank BCA sandbox returned 504 Gateway Timeout during inquiry",
		tgnotify.PerThreadID(99),
		tgnotify.WithExtra(map[string]any{"attempt": 3}),
	)
	if err != nil {
		t.Fatalf("NotifyAlert failed: %v", err)
	}

	if int(receivedPayload["message_thread_id"].(float64)) != 99 {
		t.Errorf("expected overridden thread_id 99, got %v", receivedPayload["message_thread_id"])
	}

	text := receivedPayload["text"].(string)
	if !strings.Contains(text, "BCA API Timeout") {
		t.Errorf("expected title in alert text, got: %s", text)
	}
	if !strings.Contains(text, "504 Gateway Timeout") {
		t.Errorf("expected detail in alert text, got: %s", text)
	}
}

func TestNotify_Async(t *testing.T) {
	received := make(chan bool, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- true
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true, "result": {}}`))
	}))
	defer server.Close()

	notifier, err := tgnotify.New("token", "chat123",
		tgnotify.WithBaseURL(server.URL),
		tgnotify.WithAsync(true),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = notifier.SendMessage(context.Background(), "hello async")
	if err != nil {
		t.Fatalf("SendMessage async failed: %v", err)
	}

	select {
	case <-received:
		// success
	case <-time.After(2 * time.Second):
		t.Errorf("timed out waiting for async message")
	}
}
