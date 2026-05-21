# EVM Wallet Tracker

A Go-based background service that monitors multiple EVM wallet addresses for incoming and outgoing transactions (both native ETH and ERC20 tokens) and sends notifications via Telegram. It also successfully captures internal native ETH transfers executed via smart contracts.

## Tech Stack

- **Language:** Go 1.21+
- **Ethereum Client:** `github.com/ethereum/go-ethereum` (ethclient, rpc)
- **Notifications:** Standard `net/http` for Telegram Bot API

## Installation

```bash
# Download dependencies
go mod tidy

# Build the executable
go build -o evm-tracker main.go
```

## Configuration

The application requires a JSON configuration file. Use `config.sample.json` as a template:

```bash
cp config.sample.json config.json
```

**Global Settings:**
- `rpc_ws_url`: Your EVM node WebSocket RPC URL (e.g., Alchemy, Infura). Must support `debug_traceBlockByNumber` for internal transaction tracing.
- `telegram_bot_token`: Your Telegram Bot Token.
- `telegram_chat_id`: The ID of the Telegram chat/channel to send messages to.

**Wallet Settings:**
Configure an array of `wallets` to track:
- `address`: The wallet address to monitor.
- `track_incoming` / `track_outgoing`: Booleans to toggle direction tracking.
- `track_native`: Boolean to track native ETH transfers.
- `native_min_amount`: Minimum native ETH amount to trigger an alert (in wei).
- `track_all_tokens`: Track *all* ERC20 token transfers for this wallet.
- `global_token_min_amount`: Default minimum amount for ERC20 alerts.
- `tokens`: A mapping of specific ERC20 token addresses to their custom configurations (e.g., `min_amount`).

## Running the Code

Run the service by passing the path to your config file as the first argument:

```bash
./evm-tracker config.json
```

## Architecture

The service operates as a continuous background loop:
1. **Config (`config/`)**: Loads configuration settings.
2. **State (`storage/`)**: Maintains the last processed block in `state.json` to seamlessly resume tracking after restarts.
3. **Ethereum Processing (`eth/`)**: 
   - Uses `go-ethereum` to fetch blocks and standard transactions.
   - Extracts internal native transfers using `debug_traceBlockByNumber` with the `callTracer`.
   - Filters transactions and ERC20 logs based on the per-wallet configuration logic.
4. **Notifications (`telegram/`)**: Formats matched transactions into rich HTML messages and dispatches them to Telegram.

## Development

- **Run tests:** `go test ./...`
- **Format code:** `go fmt ./...`

## Project Structure

- `main.go` - The entry point and main polling loop.
- `config/` - Configuration parsing and schemas.
- `eth/` - Core domain logic for EVM interaction, transaction matching, and internal trace extraction.
- `storage/` - State persistence.
- `telegram/` - External notification integration.

