package eth

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/user/evm-tracker/config"
)

type Processor struct {
	Wallets []config.WalletConfig
}

func NewProcessor(wallets []config.WalletConfig) *Processor {
	// Normalize configured addresses to lowercase for case-insensitive comparison
	normalizedWallets := make([]config.WalletConfig, len(wallets))
	for i, w := range wallets {
		w.Address = strings.ToLower(w.Address)
		
		// Normalize token addresses in config
		normalizedTokens := make(map[string]config.TokenConfig)
		for tAddr, tCfg := range w.Tokens {
			normalizedTokens[strings.ToLower(tAddr)] = tCfg
		}
		w.Tokens = normalizedTokens
		normalizedWallets[i] = w
	}
	return &Processor{Wallets: normalizedWallets}
}

func (p *Processor) MatchNativeTx(from common.Address, to common.Address, val *big.Int, hash string, isInternal bool) *TransactionEvent {
	fromStr := strings.ToLower(from.Hex())
	toStr := strings.ToLower(to.Hex())

	for _, w := range p.Wallets {
		if !w.TrackNative {
			continue
		}

		minAmt := new(big.Int)
		minAmt.SetString(w.NativeMinAmount, 10)

		if val.Cmp(minAmt) < 0 {
			continue
		}

		// Outgoing
		if w.TrackOutgoing && fromStr == w.Address {
			return &TransactionEvent{
				TxHash:      hash,
				IsIncoming:  false,
				IsNative:    true,
				IsInternal:  isInternal,
				Amount:      val.String(),
				FromAddress: from.Hex(),
				ToAddress:   to.Hex(),
			}
		}

		// Incoming
		if w.TrackIncoming && toStr == w.Address {
			return &TransactionEvent{
				TxHash:      hash,
				IsIncoming:  true,
				IsNative:    true,
				IsInternal:  isInternal,
				Amount:      val.String(),
				FromAddress: from.Hex(),
				ToAddress:   to.Hex(),
			}
		}
	}
	return nil
}

func (p *Processor) MatchERC20Transfer(tokenAddr, from, to common.Address, val *big.Int, hash string) *TransactionEvent {
	tokenStr := strings.ToLower(tokenAddr.Hex())
	fromStr := strings.ToLower(from.Hex())
	toStr := strings.ToLower(to.Hex())

	for _, w := range p.Wallets {
		// Check direction match first
		isOutgoing := w.TrackOutgoing && fromStr == w.Address
		isIncoming := w.TrackIncoming && toStr == w.Address
		
		if !isOutgoing && !isIncoming {
			continue
		}

		var minAmtStr string
		
		if w.TrackAllTokens {
			minAmtStr = w.GlobalTokenMinAmount
		} else {
			tokenCfg, exists := w.Tokens[tokenStr]
			if !exists {
				continue
			}
			minAmtStr = tokenCfg.MinAmount
		}

		minAmt := new(big.Int)
		if minAmtStr == "" {
			minAmtStr = "0"
		}
		minAmt.SetString(minAmtStr, 10)

		if val.Cmp(minAmt) < 0 {
			continue
		}

		return &TransactionEvent{
			TxHash:      hash,
			IsIncoming:  isIncoming,
			IsNative:    false,
			IsInternal:  false, // ERC20 transfers are just events
			Amount:      val.String(),
			TokenAddr:   tokenAddr.Hex(),
			FromAddress: from.Hex(),
			ToAddress:   to.Hex(),
		}
	}
	return nil
}
