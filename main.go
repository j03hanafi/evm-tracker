package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
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
		asset = fmt.Sprintf("ERC20 (<code>%s</code>)", evt.TokenAddr)
	}

	internalStr := ""
	if evt.IsInternal {
		internalStr = " [Internal Tx]"
	}

	return fmt.Sprintf("<b>%s %s Transfer%s</b>\nValue: %s\nFrom: <code>%s</code>\nTo: <code>%s</code>\nTxHash: <code>%s</code>",
		dir, asset, internalStr, evt.Amount, evt.FromAddress, evt.ToAddress, evt.TxHash)
}

func weiToEther(wei string) (string, error) {
	w, ok := new(big.Int).SetString(wei, 10)
	if !ok {
		return "", fmt.Errorf("bad wei: %q", wei)
	}
	// 1 ether = 1e18 wei
	div := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	f := new(big.Float).SetPrec(256).SetInt(w)
	f.Quo(f, div)
	return f.Text('f', 18), nil
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

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		log.Fatalf("get chainID: %v", err)
	}
	signer := types.LatestSignerForChainID(chainID)

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
				if ctx.Err() != nil {
					break
				}
				log.Printf("Processing block %d", b)

				// Standard Transactions
				block, err := ethClient.BlockByNumber(ctx, big.NewInt(int64(b)))
				if err != nil {
					log.Printf("Error fetching block %d: %v", b, err)
					continue // Skip saving state to retry this block later
				}

				for _, tx := range block.Transactions() {
					// Need to get sender (from address) which requires parsing the signature/chainID
					from, err := types.Sender(signer, tx)
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

				// ERC20 Logs
				erc20TransferSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
				query := ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(b)),
					ToBlock:   big.NewInt(int64(b)),
					Topics: [][]common.Hash{
						{erc20TransferSig},
					},
				}

				logs, err := ethClient.FilterLogs(ctx, query)
				if err != nil {
					log.Printf("Error fetching logs for block %d: %v", b, err)
				} else {
					for _, vLog := range logs {
						// Transfer event has 3 topics: signature, from, to
						if len(vLog.Topics) == 3 {
							from := common.BytesToAddress(vLog.Topics[1].Bytes())
							to := common.BytesToAddress(vLog.Topics[2].Bytes())
							val := new(big.Int).SetBytes(vLog.Data)

							if evt := processor.MatchERC20Transfer(vLog.Address, from, to, val, vLog.TxHash.Hex()); evt != nil {
								_ = tgClient.Send(formatMessage(evt))
							}
						}
					}
				}

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
