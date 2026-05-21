package eth

import (
	"math/big"
	"testing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/user/evm-tracker/config"
)

func TestProcessorMatchNativeTx(t *testing.T) {
	cfg := config.WalletConfig{
		Address:         "0x1111111111111111111111111111111111111111",
		TrackIncoming:   true,
		TrackOutgoing:   true,
		TrackNative:     true,
		NativeMinAmount: "1000",
	}

	p := NewProcessor([]config.WalletConfig{cfg})
	
	tests := []struct {
		name       string
		from       common.Address
		to         common.Address
		val        *big.Int
		isInternal bool
		wantMatch  bool
		wantIn     bool
	}{
		{
			name:       "outgoing native tx",
			from:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
			to:         common.HexToAddress("0x2222222222222222222222222222222222222222"),
			val:        big.NewInt(2000),
			isInternal: false,
			wantMatch:  true,
			wantIn:     false,
		},
		{
			name:       "incoming internal native tx",
			from:       common.HexToAddress("0x3333333333333333333333333333333333333333"),
			to:         common.HexToAddress("0x1111111111111111111111111111111111111111"),
			val:        big.NewInt(2000),
			isInternal: true,
			wantMatch:  true,
			wantIn:     true,
		},
		{
			name:       "below min amount",
			from:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
			to:         common.HexToAddress("0x2222222222222222222222222222222222222222"),
			val:        big.NewInt(500),
			isInternal: false,
			wantMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := p.MatchNativeTx(tt.from, tt.to, tt.val, "0xabc", tt.isInternal)
			if tt.wantMatch && evt == nil {
				t.Fatalf("expected match, got nil")
			}
			if !tt.wantMatch && evt != nil {
				t.Fatalf("expected no match, got %v", evt)
			}
			if tt.wantMatch {
				if evt.IsIncoming != tt.wantIn {
					t.Errorf("IsIncoming: got %v, want %v", evt.IsIncoming, tt.wantIn)
				}
				if evt.IsInternal != tt.isInternal {
					t.Errorf("IsInternal: got %v, want %v", evt.IsInternal, tt.isInternal)
				}
			}
		})
	}
}

func TestProcessorMatchERC20Transfer(t *testing.T) {
	cfg := config.WalletConfig{
		Address:              "0x1111111111111111111111111111111111111111",
		TrackIncoming:        true,
		TrackOutgoing:        true,
		TrackAllTokens:       false,
		GlobalTokenMinAmount: "0",
		Tokens: map[string]config.TokenConfig{
			"0x4444444444444444444444444444444444444444": {MinAmount: "500"},
		},
	}
	p := NewProcessor([]config.WalletConfig{cfg})

	tests := []struct {
		name      string
		token     common.Address
		from      common.Address
		to        common.Address
		val       *big.Int
		wantMatch bool
	}{
		{
			name:      "valid tracked token transfer",
			token:     common.HexToAddress("0x4444444444444444444444444444444444444444"),
			from:      common.HexToAddress("0x1111111111111111111111111111111111111111"),
			to:        common.HexToAddress("0x2222222222222222222222222222222222222222"),
			val:       big.NewInt(1000),
			wantMatch: true,
		},
		{
			name:      "untracked token",
			token:     common.HexToAddress("0x5555555555555555555555555555555555555555"),
			from:      common.HexToAddress("0x1111111111111111111111111111111111111111"),
			to:        common.HexToAddress("0x2222222222222222222222222222222222222222"),
			val:       big.NewInt(1000),
			wantMatch: false,
		},
		{
			name:      "below min amount",
			token:     common.HexToAddress("0x4444444444444444444444444444444444444444"),
			from:      common.HexToAddress("0x1111111111111111111111111111111111111111"),
			to:        common.HexToAddress("0x2222222222222222222222222222222222222222"),
			val:       big.NewInt(100),
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := p.MatchERC20Transfer(tt.token, tt.from, tt.to, tt.val, "0xabc")
			if tt.wantMatch && evt == nil {
				t.Fatalf("expected match, got nil")
			}
			if !tt.wantMatch && evt != nil {
				t.Fatalf("expected no match, got %v", evt)
			}
		})
	}
}
