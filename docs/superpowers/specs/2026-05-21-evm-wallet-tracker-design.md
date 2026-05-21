# EVM Wallet Tracker Design

## Purpose
A Go-based service that monitors multiple EVM wallet addresses for incoming and/or outgoing transactions (both native ETH and specific/all ERC20 tokens) based on per-wallet configuration, and sends notifications via Telegram.

## Architecture
The application will be built using a modular Go architecture:

### 1. Configuration (`config` package)
- Reads configuration from a local `config.json` file.
- Global Configuration includes:
  - `rpc_ws_url`: The WebSocket URL for the EVM RPC node (e.g., Alchemy/Infura).
  - `telegram_bot_token`: The token for the Telegram bot.
  - `telegram_chat_id`: The ID of the chat/channel to send notifications to.
- Per-Wallet Configuration (Array of objects):
  - `address`: The wallet address to monitor.
  - `track_incoming` (bool): Whether to track incoming transfers.
  - `track_outgoing` (bool): Whether to track outgoing transfers.
  - `track_native` (bool): Whether to track native ETH transfers.
  - `native_min_amount` (string): Minimum amount of native ETH to trigger a notification (in wei).
  - `track_all_tokens` (bool): Whether to track all ERC20 tokens or only a specific list.
  - `global_token_min_amount` (string, optional): Default minimum amount for tokens to trigger a notification.
  - `tokens` (map[string]TokenConfig): A dictionary of specific ERC20 token addresses to their config (e.g., `{"0xUSDC": {"min_amount": "1000000"}}`).

### 2. State Management (`storage` package)
- Persists the last processed block to a local `state.json` file.
- The file will only store the absolute latest processed block globally (e.g., `{"latest_block": 12345678}`) and will be completely overwritten on each update to maintain a minimal file size.
- On startup, the app reads this file to resume monitoring from the last processed block, preventing missed transactions during downtime.

### 3. Ethereum Integration (`eth` package)
- Uses the `github.com/ethereum/go-ethereum` library.
- Connects to the EVM node via WebSocket.
- Subscribes to new block headers (`SubscribeNewHead`).
- Upon receiving a new block header, or when catching up from the saved state:
  - Fetches the full block by number.
  - Processes all transactions and receipts in the block.
  - **Native ETH Internal Transactions**: Standard block data does not include internal value transfers (ETH sent via smart contract execution). The app will use `debug_traceBlockByNumber` (or similar tracing API) to extract internal call traces, ensuring no native ETH transfers via smart contracts are missed.
  - Checks Native ETH transfers (from both standard Txs and internal traces): if sender/receiver matches a configured wallet, and `track_native` is true for that direction, and amount >= `native_min_amount`.
  - **ERC20 Internal Transactions**: Unlike native ETH, ERC20 token transfers executed internally by smart contracts *always* emit standard `Transfer` events. These are naturally captured in the standard transaction receipt logs, so no special tracing is needed for them.
  - Checks ERC20 `Transfer` logs (capturing both direct and internal smart contract transfers): if sender/receiver matches a configured wallet, and direction is tracked.
    - If `track_all_tokens` is true, check amount >= `global_token_min_amount`.
    - If `track_all_tokens` is false, check if the contract address is in the wallet's `tokens` map, and amount >= the specific token's `min_amount`.
- Emits detected valid transactions to a Go channel for processing by the notification service.

### 4. Notifications (`telegram` package)
- Receives detected transactions.
- Formats the transaction details into a readable Markdown/HTML message (including Type (In/Out), Tx Hash, Value, Token Symbol/Name (if possible, or address), Sender, Receiver, and a flag if it was an internal transfer).
- Sends the message using the Telegram Bot API via HTTP POST requests.

### 5. Application Core (`main` package)
- Initializes the configuration, storage, Ethereum client, and Telegram client.
- Wires the components together.
- Implements graceful shutdown handling (listening for SIGINT/SIGTERM) to ensure the final processed block is saved to `state.json` before exiting.

## Error Handling
- **RPC Disconnects**: Implement reconnection logic if the WebSocket connection drops.
- **Telegram API Errors**: Log errors and optionally retry sending the notification.
