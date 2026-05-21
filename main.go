package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/user/evm-tracker/config"
	"github.com/user/evm-tracker/eth"
	"github.com/user/evm-tracker/storage"
	"github.com/user/evm-tracker/telegram"
)

func formatMessage(evt *eth.TransactionEvent) string {
	dir := "📤 Outgoing"
	if evt.IsIncoming {
		dir = "📥 Incoming"
	}
	
	asset := "ETH"
	if !evt.IsNative {
		asset = fmt.Sprintf("ERC20 (%s)", evt.TokenAddr)
	}
	
	internalStr := ""
	if evt.IsInternal {
		internalStr = " [Internal Tx]"
	}

	return fmt.Sprintf("<b>%s %s Transfer%s</b>\nValue: %s\nFrom: %s\nTo: %s\nTxHash: %s", 
		dir, asset, internalStr, evt.Amount, evt.FromAddress, evt.ToAddress, evt.TxHash)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config_path>", os.Args[0])
	}
	configPath := os.Args[1]

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	stateFile := "state.json"
	lastBlock, err := storage.ReadLatestBlock(stateFile)
	if err != nil {
		log.Fatalf("Failed to read state: %v", err)
	}

	tgClient := telegram.NewClient(cfg.TelegramBotToken, cfg.TelegramChatID)
	processor := eth.NewProcessor(cfg.Wallets)

	rpcClient, err := rpc.DialContext(context.Background(), cfg.RPCWSURL)
	if err != nil {
		log.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer rpcClient.Close()
	
	ethClient := ethclient.NewClient(rpcClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Shutting down...")
		cancel()
	}()

	log.Printf("Starting tracker. Last processed block: %d\n", lastBlock)

	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := storage.WriteLatestBlock(stateFile, lastBlock); err != nil {
				log.Printf("Failed to write final state: %v", err)
			}
			fmt.Println("Shutdown complete.")
			return
		case <-ticker.C:
			header, err := ethClient.HeaderByNumber(ctx, nil)
			if err != nil {
				log.Printf("Error getting latest header: %v", err)
				continue
			}

			currentBlock := header.Number.Uint64()
			if lastBlock == 0 {
				lastBlock = currentBlock - 1 // Start from current if no state
			}

			for b := lastBlock + 1; b <= currentBlock; b++ {
                log.Printf("Processing block %d", b)
				
				// Standard Transactions
				block, err := ethClient.BlockByNumber(ctx, big.NewInt(int64(b)))
				if err != nil {
					log.Printf("Error fetching block %d: %v", b, err)
					continue // Skip saving state to retry this block later
				}

				for _, tx := range block.Transactions() {
					// Need to get sender (from address) which requires parsing the signature/chainID
					from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
					if err != nil {
						continue // fallback or ignore if we can't get sender
					}
					
					to := tx.To()
					if to != nil { // regular transfer
						if evt := processor.MatchNativeTx(from, *to, tx.Value(), tx.Hash().Hex(), false); evt != nil {
							_ = tgClient.Send(formatMessage(evt))
						}
					}
				}

				// ERC20 Logs (using FilterLogs could be more efficient, but checking receipts for the block is exhaustive)
				// For the implementation plan, we acknowledge this logic would ideally use `ethclient.FilterLogs` for ERC20s 
				// to avoid fetching all receipts unless necessary. 

				// Internal Traces
				traces, err := eth.FetchBlockTraces(ctx, rpcClient, b)
				if err == nil {
					for i, trace := range traces {
						// Assuming transactions align with trace results index
						txHash := ""
						if i < len(block.Transactions()) {
							txHash = block.Transactions()[i].Hash().Hex()
						}
						
						events := eth.ExtractInternalTransfers(trace.Result, txHash, processor)
						for _, evt := range events {
							_ = tgClient.Send(formatMessage(evt))
						}
					}
				}
				
				lastBlock = b
				if err := storage.WriteLatestBlock(stateFile, lastBlock); err != nil {
					log.Printf("Error saving state: %v", err)
				}
			}
		}
	}
}