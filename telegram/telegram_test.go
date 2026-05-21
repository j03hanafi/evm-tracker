package telegram

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendNotification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botTEST_TOKEN/sendMessage" {
			t.Errorf("Expected path /botTEST_TOKEN/sendMessage, got %s", r.URL.Path)
		}
		
		r.ParseForm()
		chatId := r.FormValue("chat_id")
		if chatId != "TEST_CHAT" {
			t.Errorf("Expected chat_id 'TEST_CHAT', got %s", chatId)
		}
		
		text := r.FormValue("text")
		if text != "Test Message" {
			t.Errorf("Expected text 'Test Message', got %s", text)
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("TEST_TOKEN", "TEST_CHAT")
	client.ApiURL = server.URL // Override API URL for testing

	err := client.Send("Test Message")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}
