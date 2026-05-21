# Project Navigation (for LLM)

> Short index for working in this codebase. Full details: see `README.md`.

## Stack
Go 1.21+ / go-ethereum / net/http

## Entry Point
- Type: Background Service / CLI
- File: `main.go`

## Architecture Pattern
Modular Package Design / Continuous Worker Loop

## Where Things Live
| To work on... | Open... |
|---------------|---------|
| Entry Point & Event Loop | `main.go` |
| ETH Transaction Matching Logic | `eth/processor.go` |
| Internal Trace Extraction | `eth/client.go` |
| Data Models | `eth/types.go`, `config/config.go` |
| Telegram API Integration | `telegram/telegram.go` |
| Block State Persistence | `storage/state.go` |
| Tests | `*_test.go` in respective packages |

## Traversal Hint
When investigating a feature, start from the continuous polling loop in `main.go` -> standard tx parsing or `eth.FetchBlockTraces` -> `eth.Processor` logic -> `telegram.Send`.

## Commands
- Run: `./evm-tracker <config.json>`
- Test: `go test ./...`
- Build: `go build -o evm-tracker main.go`

## Conventions
- **Naming:** Idiomatic Go (short, descriptive package names).
- **Errors:** Standard Go `fmt.Errorf("...: %w", err)` wrapping.
- **Tests:** Table-driven tests utilizing `t.Run()` subtests (see `eth/processor_test.go`).

