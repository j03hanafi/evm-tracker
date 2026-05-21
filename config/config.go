package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type TokenConfig struct {
	MinAmount string `json:"min_amount"`
}

type WalletConfig struct {
	Address               string                 `json:"address"`
	TrackIncoming         bool                   `json:"track_incoming"`
	TrackOutgoing         bool                   `json:"track_outgoing"`
	TrackNative           bool                   `json:"track_native"`
	NativeMinAmount       string                 `json:"native_min_amount"`
	TrackAllTokens        bool                   `json:"track_all_tokens"`
	GlobalTokenMinAmount  string                 `json:"global_token_min_amount"`
	Tokens                map[string]TokenConfig `json:"tokens"`
}

type AppConfig struct {
	RPCWSURL           string         `json:"rpc_ws_url"`
	TelegramBotToken   string         `json:"telegram_bot_token"`
	TelegramChatID     string         `json:"telegram_chat_id"`
	Wallets            []WalletConfig `json:"wallets"`
}

func Load(filename string) (*AppConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
