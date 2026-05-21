package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	testJSON := `{
		"rpc_ws_url": "wss://test",
		"telegram_bot_token": "token123",
		"telegram_chat_id": "chat123",
		"wallets": [
			{
				"address": "0x123",
				"track_incoming": true,
				"track_outgoing": true,
				"track_native": true,
				"native_min_amount": "100",
				"track_all_tokens": false,
				"global_token_min_amount": "0",
				"tokens": {
					"0xUSDC": {"min_amount": "1000"}
				}
			}
		]
	}`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.json")
	
	err := os.WriteFile(configFile, []byte(testJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.RPCWSURL != "wss://test" {
		t.Errorf("Expected rpc_ws_url 'wss://test', got '%s'", cfg.RPCWSURL)
	}
	if len(cfg.Wallets) != 1 {
		t.Fatalf("Expected 1 wallet, got %d", len(cfg.Wallets))
	}
	if cfg.Wallets[0].Address != "0x123" {
		t.Errorf("Expected wallet address '0x123', got '%s'", cfg.Wallets[0].Address)
	}
	if cfg.Wallets[0].Tokens["0xUSDC"].MinAmount != "1000" {
		t.Errorf("Expected token min_amount '1000', got '%s'", cfg.Wallets[0].Tokens["0xUSDC"].MinAmount)
	}
}
